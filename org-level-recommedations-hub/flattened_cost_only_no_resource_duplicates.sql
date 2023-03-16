SELECT
  agg.table.*
    FROM (
      select target_resource,
        ARRAY_AGG(STRUCT(table)
          ORDER BY  
            impact_cost_unit DESC)[SAFE_OFFSET(0)] agg
      FROM
        `recommender-export-377819.recommendations_export_dataset.flattened_recommendations` table cross join unnest(target_resources) as target_resource
      where has_impact_cost = true AND recommender_subtype in ('CHANGE MACHINE TYPE', 'STOP VM')
      GROUP BY target_resource)
union all (
  SELECT * 
  from `recommender-export-377819.recommendations_export_dataset.flattened_recommendations`
  where has_impact_cost = true AND recommender_subtype NOT IN ('CHANGE MACHINE TYPE', 'STOP VM'))