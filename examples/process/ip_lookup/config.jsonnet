local ipdb = import '../../../build/config/ip_database.libsonnet';
local process = import '../../../build/config/process.libsonnet';

// the MaxMind City database can be read from local disk, HTTP(S) URL, or AWS S3 URL
// other databases / providers can be used by changing the imported function (e.g., ipdb.maxmind_asn)
local mm_city = ipdb.maxmind_city('location://path/to/maxmind.mmdb');

// applies the IPDatabase processor using a MaxMind City database
process.ip_database(input='addr', output='geo', database_options=mm_city)
