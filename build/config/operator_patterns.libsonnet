local operator = import 'operator.libsonnet';

{
  and(inspectors): operator.operate(operator='and', inspectors=inspectors),
  or(inspectors): operator.operate(operator='or', inspectors=inspectors),
  nand(inspectors): operator.operate(operator='nand', inspectors=inspectors),
  nor(inspectors): operator.operate(operator='nor', inspectors=inspectors),
}
