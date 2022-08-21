local processlib = import '../../../build/config/process.libsonnet';
local conditionlib = import '../../../build/config/condition.libsonnet';

// applies the Insert processor if any of these conditions match
local conditions = [
	conditionlib.strings.equals(key='foo', expression='bar'),
	conditionlib.strings.equals(key='baz', expression='qux')
];

processlib.insert(output='xyzzy', value='thud', condition_operator='or', condition_inspectors=conditions)
