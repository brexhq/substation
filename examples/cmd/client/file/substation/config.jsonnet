local event = import 'event.libsonnet';
local send = import 'send.libsonnet';

{
  transforms:
    event.transforms
    + send.transforms,
}
