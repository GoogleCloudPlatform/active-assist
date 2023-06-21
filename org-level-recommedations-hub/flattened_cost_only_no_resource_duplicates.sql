SELECT
  agg.table.*
    FROM (
      select target_resource,
        ARRAY_AGG(STRUCT(table)
          ORDER BY  
            impact_cost_unit DESC)[SAFE_OFFSET(0)] agg
      FROM
        `${var.project_id}.${google_bigquery_dataset.org_level_rec_hub_dataset.dataset_id}.${google_bigquery_table.flattened_recommendations.table_id}` table cross join unnest(target_resources) as target_resource
      where has_impact_cost = true AND recommender_subtype in ('CHANGE MACHINE TYPE', 'STOP VM')
      GROUP BY target_resource)
union all (
  SELECT * 
  from `${var.project_id}.${google_bigquery_dataset.org_level_rec_hub_dataset.dataset_id}.${google_bigquery_table.flattened_recommendations.table_id}`
  where has_impact_cost = true AND recommender_subtype NOT IN ('CHANGE MACHINE TYPE', 'STOP VM'))