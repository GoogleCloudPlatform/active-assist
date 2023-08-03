[
  {"description":"Represents what cloud entity type the recommendation was generated for - eg: project number, billing account\\n","mode":"NULLABLE","name":"cloud_entity_type","type":"STRING"},
  {"description":"Value of the project number or billing account id\\n","mode":"NULLABLE","name":"cloud_entity_id","type":"STRING"},
  {"description":"Name of recommendation. A project recommendation is represented as\\nprojects/[PROJECT_NUMBER]/locations/[LOCATION]/recommenders/[RECOMMENDER_ID]/recommendations/[RECOMMENDATION_ID]\\n","mode":"NULLABLE","name":"name","type":"STRING"},
  {"description":"Location for which this recommendation is generated\\n","mode":"NULLABLE","name":"location","type":"STRING"},
  {"description":"Recommender ID of the recommender that has produced this recommendation\\n","mode":"NULLABLE","name":"recommender","type":"STRING"},
  {"description":"Contains an identifier for a subtype of recommendations produced for the\\nsame recommender. Subtype is a function of content and impact, meaning a\\nnew subtype will be added when either content or primary impact category\\nchanges.\\nExamples:\\nFor recommender = \"google.iam.policy.Recommender\",\\nrecommender_subtype can be one of \"REMOVE_ROLE\"/\"REPLACE_ROLE\"\\n","mode":"NULLABLE","name":"recommender_subtype","type":"STRING"},
  {"description":"Contains the fully qualified resource names for resources changed by the\\noperations in this recommendation. This field is always populated. ex:\\n[//cloudresourcemanager.googleapis.com/projects/foo].\\n","mode":"REPEATED","name":"target_resources","type":"STRING"},
  {"description":"Required. Free-form human readable summary in English.\\nThe maximum length is 500 characters.\\n","mode":"NULLABLE","name":"description","type":"STRING"},
  {"description":"Output only. Last time this recommendation was refreshed by the system that created it in the first place.\\n","mode":"NULLABLE","name":"last_refresh_time","type":"TIMESTAMP"},
  {"description":"Required. The primary impact that this recommendation can have while trying to optimize\\nfor one category.\\n","fields":[
    {"description":"Category that is being targeted.\\nValues can be the following:\\n  CATEGORY_UNSPECIFIED:\\n    Default unspecified category. Don't use directly.\\n  COST:\\n    Indicates a potential increase or decrease in cost.\\n  SECURITY:\\n    Indicates a potential increase or decrease in security.\\n  PERFORMANCE:\\n    Indicates a potential increase or decrease in performance.\\n","mode":"NULLABLE","name":"category","type":"STRING"},
    {"description":"Optional. Use with CategoryType.COST","fields":[
      {"description":"An approximate projection on amount saved or amount incurred.\\nNegative cost units indicate cost savings and positive cost units indicate\\nincrease. See google.type.Money documentation for positive/negative units.\\n","fields":[
        {"description":"The 3-letter currency code defined in ISO 4217.","mode":"NULLABLE","name":"currency_code","type":"STRING"},
        {"description":"The whole units of the amount. For example if `currencyCode` is `\"USD\"`,\\nthen 1 unit is one US dollar.\\n","mode":"NULLABLE","name":"units","type":"INTEGER"},
        {"description":"Number of nano (10^-9) units of the amount.\\nThe value must be between -999,999,999 and +999,999,999 inclusive.\\nIf `units` is positive, `nanos` must be positive or zero.\\nIf `units` is zero, `nanos` can be positive, zero, or negative.\\nIf `units` is negative, `nanos` must be negative or zero.\\nFor example $-1.75 is represented as `units`=-1 and `nanos`=-750,000,000.\\n","mode":"NULLABLE","name":"nanos","type":"INTEGER"}
      ],"mode":"NULLABLE","name":"cost","type":"RECORD"},
      {"description":"Duration for which this cost applies.","fields":[
        {"description":"Signed seconds of the span of time. Must be from -315,576,000,000\\nto +315,576,000,000 inclusive. Note: these bounds are computed from:\\n60 sec/min * 60 min/hr * 24 hr/day * 365.25 days/year * 10000 years\\n","mode":"NULLABLE","name":"seconds","type":"INTEGER"},
        {"description":"Signed fractions of a second at nanosecond resolution of the span\\nof time. Durations less than one second are represented with a 0\\n`seconds` field and a positive or negative `nanos` field. For durations\\nof one second or more, a non-zero value for the `nanos` field must be\\nof the same sign as the `seconds` field. Must be from -999,999,999\\nto +999,999,999 inclusive.\\n","mode":"NULLABLE","name":"nanos","type":"INTEGER"}
      ],"mode":"NULLABLE","name":"duration","type":"RECORD"}
    ],"mode":"NULLABLE","name":"cost_projection","type":"RECORD"}
  ],"mode":"NULLABLE","name":"primary_impact","type":"RECORD"},
  {"description":"Output only. The state of the recommendation:\\n  STATE_UNSPECIFIED:\\n    Default state. Don't use directly.\\n  ACTIVE:\\n    Recommendation is active and can be applied. Recommendations content can\\n    be updated by Google.\\n    ACTIVE recommendations can be marked as CLAIMED, SUCCEEDED, or FAILED.\\n  CLAIMED:\\n    Recommendation is in claimed state. Recommendations content is\\n    immutable and cannot be updated by Google.\\n    CLAIMED recommendations can be marked as CLAIMED, SUCCEEDED, or FAILED.\\n  SUCCEEDED:\\n    Recommendation is in succeeded state. Recommendations content is\\n    immutable and cannot be updated by Google.\\n    SUCCEEDED recommendations can be marked as SUCCEEDED, or FAILED.\\n  FAILED:\\n    Recommendation is in failed state. Recommendations content is immutable\\n    and cannot be updated by Google.\\n    FAILED recommendations can be marked as SUCCEEDED, or FAILED.\\n  DISMISSED:\\n    Recommendation is in dismissed state.\\n    DISMISSED recommendations can be marked as ACTIVE.\\n","mode":"NULLABLE","name":"state","type":"STRING"},
  {"description":"Ancestry for the recommendation entity\\n","fields":[
    {"description":"Organization to which the recommendation project\\n","mode":"NULLABLE","name":"organization_id","type":"STRING"},
    {"description":"Up to 5 levels of parent folders for the recommendation project\\n","mode":"REPEATED","name":"folder_ids","type":"STRING"}
  ],"mode":"NULLABLE","name":"ancestors","type":"RECORD"},
  {"description":"Insights associated with this recommendation. A project insight is represented as\\nprojects/[PROJECT_NUMBER]/locations/[LOCATION]/insightTypes/[INSIGHT_TYPE_ID]/insights/[insight_id]\\n","mode":"REPEATED","name":"associated_insights","type":"STRING"},
  {"description":"Additional details about the recommendation in JSON format\\n","mode":"NULLABLE","name":"recommendation_details","type":"STRING"},
  {"description":"Priority of the recommendation:\\n  PRIORITY_UNSPECIFIED:\\n    Default unspecified priority. Don't use directly.\\n  P4:\\n    Lowest priority.\\n  P3:\\n    Second lowest priority.\\n  P2:\\n    Second highest priority.\\n  P1:\\n    Highest priority.\\n","mode":"NULLABLE","name":"priority","type":"STRING"}
]

