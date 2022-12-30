local operator = import 'operator.libsonnet';

{
  all(inspectors): operator.operate(operator='all', inspectors=inspectors),
  any(inspectors): operator.operate(operator='any', inspectors=inspectors),
  none(inspectors): operator.operate(operator='none', inspectors=inspectors),
}
