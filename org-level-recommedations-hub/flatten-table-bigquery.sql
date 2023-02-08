# This is a bigquery script designed to be used in Data Studio to properly display the data.
# Keep in mind this script is currently under development, and probably is not optimized or in a final state. 
# The reason this script was created is because you need to flatten individual recommendations to single rows if you want to make dashboards in Datastudio efficiently
# I also am selecting individual columns to help increase the speed in Datastudio. 
# Known issues:
# Currently we dont handle additional currencies well.
# Folder_ids do not map to folder names (missing from asset inventory)
# Some project names are missing
## *** Final Select ***
select 
  # These are the final columns used in Datastudio
  project_name,
  project_id,
  asset_type,
  name as recommender_name,
  location,
  REPLACE(recommender_subtype, "_", " ") as recommender_subtype,
  # If we dont use distinct here, sum calculations could get messed up in datastudio. Applies to all array_aggs in this query. 
  ARRAY_AGG(distinct target_resource) as target_resources,
  last_refresh_time as recommender_last_refresh_time,
  primary_impact.category as impact_category,
  ABS(primary_impact.cost_projection.cost.units) as impact_cost_unit,
  primary_impact.cost_projection.cost.currency_code as impact_currency_code,
  state as recommender_state,
  ARRAY_AGG(distinct folder_id ignore nulls) as folder_ids,
  ARRAY_AGG(distinct insight_id) as insight_ids,
  ARRAY_AGG(STRUCT(insight_name, insight_type, insight_subtype, category as insight_category, insight_state)) as insights
  ## *** Join Query to build initial table *** 
  from (
    select * except(associated_insights, target_resources)
    # We just want to grab the latest refresh time per export
    from 
    ## *** Query to grab only the LATEST recommendations ***
    (SELECT
        agg.table.*
      FROM (
        SELECT
          name,
          # Small note for anyone who comes after me
          # ARRAY_AGGs are cool because you can order the results,
          # and then grab as many of the results as you want. 
          # In this case we are grabbing on the newest result. 
          ARRAY_AGG(STRUCT(table)
          ORDER BY
            last_refresh_time DESC)[SAFE_OFFSET(0)] agg
        FROM
          `finOps.recommendations_export` table
        GROUP BY
          name)
    ) as r
    # Need to unnest fields for joins
    cross join unnest(associated_insights) as insight_id
    cross join unnest(target_resources) as target_resource
    # Cross joins will remove nulls, and in our case we still need nulls
    left join unnest(ancestors.folder_ids) as folder_id
    ### Join Cloud Assets
    left join  (
      select 
      REGEXP_EXTRACT(name, r'/([^/]+)/?$') as project_name,
      REGEXP_EXTRACT(ancestor,  r'/([^/]+)/?$') as project_id,
      asset_type from 
      (select * from `finOps.cloudAssets` 
      cross join unnest(ancestors) as ancestor
      where asset_type in ("compute.googleapis.com/Project")
      and ancestor like "projects/%")
    ) as a
    on r.cloud_entity_id = a.project_id
    ### Join insights
    # Currently there is a bug where VM resizing recommendations are missing associated insights. This should be fixed soon. 
    left join (select 
      name as insight_name,
      insight_type,
      insight_subtype,
      category,
      state as insight_state,
      a_r
      from finOps.insights_export
      cross join unnest(associated_recommendations) as a_r
    ) as i
    on r.name=i.a_r
  )
  group by 1,2,3,4,5,6,8,9,10,11,12
  order by recommender_name