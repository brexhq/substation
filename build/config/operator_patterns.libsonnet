local condition = import 'condition.libsonnet';

{
  and(inspectors): condition.operator(operator='and', inspectors=inspectors),
  or(inspectors): condition.operator(operator='or', inspectors=inspectors),
  nand(inspectors): condition.operator(operator='nand', inspectors=inspectors),
  nor(inspectors): condition.operator(operator='nor', inspectors=inspectors),
}
