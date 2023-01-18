local sub = import '../../../../../build/config/substation.libsonnet';

// calls the DNS processor if the value in 'addr' is a public
// IP address.
local dns = sub.interfaces.processor.dns(
  settings={
    key: 'addr',
    set_key: 'domain',
    condition: sub.patterns.operator.ip.public(key='addr'),
  },
  options={ type: 'reverse_lookup' }
);

// an in-memory read-write key-value (KV) store is used to 
// cache DNS requests. this cache is active for the lifecycle 
// of the Lambda (approximately 2 hours). 
local kv = sub.interfaces.kv_store.memory(
  settings={
    capacity: 1024,
  }
);

local processors = [
  // DNS processor is wrapped in the KV store -- DNS requests
  // are only made if the input is not already in the store.
  // values are kept in the store for 1 hour.
  sub.patterns.processor.kv_store.cache_aside(
    processor=dns, kv_options=kv, offset_ttl=60 * 60, prefix='dns'
  ),
  // the enriched field is moved and copied so that only the enriched
  // data is returned in a new object, otherwise the entire object is
  // returned to the caller.
  sub.patterns.processor.move(key='@this', set_key='!metadata this'),
  sub.interfaces.processor.copy(
    settings={key: '!metadata this.domain', set_key: 'domain'}
  ),
];

{
  processors: sub.helpers.flatten_processors(processors),
}
