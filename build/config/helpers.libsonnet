{
  // if input is not an array, then this returns an array
  make_array(i): if !std.isArray(i) then [i] else i,
  key: {
    // if key is foo and arr is bar, then result is foo.bar
    // if key is foo and arr is [bar, baz], then result is foo.bar.baz
    append(key, arr): std.join('.', $.make_array(key) + $.make_array(arr)),
    // if key is foo, then result is foo.-1
    append_array(key): key + '.-1',
    // if key is foo and e is 0, then result is foo.0
    get_element(key, e=0): std.join('.', [key, if std.isNumber(e) then std.toString(e) else e]),
  },
  // dynamically flattens processor configurations
  flatten_processors(processors): std.flattenArrays([
    if std.objectHas(p, 'processors') then p.processors
    else if std.objectHas(p, 'processor') then [p.processor]
    else [p]

    for p in $.make_array(processors)
  ]),
}
