local sub = import '../../../../../../../build/config/substation.libsonnet';
local const = import '../const.libsonnet';

{
  concurrency: 1,
  transforms: [
    // The user's status is determined to be inactive if there is a successful deletion event.
    // Any other successful authentication event will set the user's status to active.
    //
    // In production deployments, additional filtering should be used to reduce the number of
    // queries made to the KV store.
    sub.tf.meta.switch({ cases: [
      {
        condition: sub.cnd.all([
          sub.cnd.str.eq({ object: { source_key: 'event.category' }, value: 'authentication' }),
          sub.cnd.str.eq({ object: { source_key: 'event.type' }, value: 'deletion' }),
          sub.cnd.str.eq({ object: { source_key: 'event.outcome' }, value: 'success' }),
        ]),
        transform: sub.tf.object.insert({ object: { target_key: 'user.status.-1' }, value: 'idp_inactive' }),
      },
      {
        condition: sub.cnd.all([
          sub.cnd.str.eq({ object: { source_key: 'event.outcome' }, value: 'success' }),
        ]),
        transform: sub.tf.object.insert({ object: { target_key: 'user.status.-1' }, value: 'idp_active' }),
      },
    ] }),
    // Puts the user's metadata into the KV store indexed by their email address.
    sub.tf.enrich.kv_store.iset({ obj: { src: 'user.email', trg: 'user' }, prefix: 'idp_user', kv_store: const.kv_store }),
  ],
}
