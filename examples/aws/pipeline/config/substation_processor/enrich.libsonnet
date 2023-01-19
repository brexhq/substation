local sub = import '../../../../../build/config/substation.libsonnet';

local processors = [
  // calls an enrichment node which is running as a microservice.
  // this node performs DNS resolution against the 'addr' field. 
  sub.interfaces.processor.aws_lambda(
    options={function_name: 'substation_lambda_enrichment'},
    settings={key:'@this', set_key: 'lambda', ignore_errors: true, condition: sub.interfaces.operator.all([
      sub.patterns.inspector.length.gt_zero(key='addr'),
    ])},
  ),
];

{
  processors: sub.helpers.flatten_processors(processors),
}
