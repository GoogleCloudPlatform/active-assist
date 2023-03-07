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
  project_name,
  project_id,
  asset_type,
  name as recommender_name,
  location,
  REPLACE(recommender_subtype, "_", " ") as recommender_subtype,
  ARRAY_AGG(distinct target_resource) as target_resources,
  last_refresh_time as recommender_last_refresh_time,
  primary_impact.category as impact_category,
  if(primary_impact.cost_projection.cost.units is null,false, true) as has_impact_cost,
  ABS(primary_impact.cost_projection.cost.units) as impact_cost_unit,
  primary_impact.cost_projection.cost.currency_code as impact_currency_code,
  state as recommender_state,
  ARRAY_AGG(distinct folder_id ignore nulls) as folder_ids,
  ARRAY_AGG(distinct insight_id) as insight_ids,
  ARRAY_AGG(STRUCT(insight_name, insight_type, insight_subtype, category as insight_category, insight_state)) as insights
  from (
    select * except(associated_insights, target_resources)
    # We just want to grab the latest refresh time per export
    from (SELECT
        agg.table.*
      FROM (
        select target_resource,
        recommender_subtype,
            ARRAY_AGG(STRUCT(table)
              ORDER BY  
                last_refresh_time DESC)[SAFE_OFFSET(0)] agg
        FROM
          `recommender-export-377819.recommendations_export_dataset.recommendations_export` table cross join unnest(target_resources) as target_resource
          GROUP BY target_resource, recommender_subtype)
    ) as r
    cross join unnest(associated_insights) as insight_id
    cross join unnest(target_resources) as target_resource
    #Cross join will remove nulls, and in our case we still need nulls
    left join unnest(ancestors.folder_ids) as folder_id
    left join  (
      select 
      REGEXP_EXTRACT(name, r'/([^/]+)/?$') as project_name,
      REGEXP_EXTRACT(ancestor,  r'/([^/]+)/?$') as project_id,
      asset_type from 
      (select * from `recommender-export-377819.recommendations_export_dataset.asset_export_table`
      cross join unnest(ancestors) as ancestor
      where asset_type in ("compute.googleapis.com/Project")
      and ancestor like "projects/%")
    ) as a
    on r.cloud_entity_id = a.project_id
    left join (select 
      name as insight_name,
      insight_type,
      insight_subtype,
      category,
      state as insight_state,
      a_r
      from `recommender-export-377819.recommendations_export_dataset.insights_export`
      cross join unnest(associated_recommendations) as a_r
    ) as i
    on r.name=i.a_r
  )
  group by 1,2,3,4,5,6,8,9,10,11,12,13
  order by recommender_name