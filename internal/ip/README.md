# ip

Contains tools for modifying and enriching IP address data. Enriched IP address information is abstracted from the source into one of many shared data structures:

* location: contains geolocation information
* asn: contains autonomous system information

The package provides read access to enrichment databases that are setup using environment variables. Each environment variable should contain the location of the database, which can be either a path on local disk, an HTTP(S) URL, or an AWS S3 URL. 

Reading each database is achieved using these methods:

* ASN: returns autonomous system information for an IP address
* Location: returns geolocation information for an IP address

These databases (along with their setup environment variables) are supported:

* IP2Location (IP2LOCATION_DB)
* MaxMind GeoIP2 City (MAXMIND_LOCATION_DB)
* MaxMind GeoLite2 City (MAXMIND_LOCATION_DB)
* MaxMind GeoLite2 ASN (MAXMIND_ASN_DB)
