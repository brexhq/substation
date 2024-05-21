// This example shows how to use the `utility_control` transform to
// generate a control (ctrl) Message based on the amount of data Messages
// received by the system. ctrl Messages overrides the settings of the
// `aggregate_to_array` transform (and any other transform that supports).
local sub = import '../../../../../build/config/substation.libsonnet';

{
  transforms: [
    sub.tf.utility.control({ batch: { count: 2 } }),
    sub.tf.aggregate.to.array({ batch: { count: 10000 } }),
    sub.tf.send.stdout(),
  ],
}
