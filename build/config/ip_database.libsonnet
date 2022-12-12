{
  ip2location(database): {
    type: 'ip2location',
    settings: { database: database },
  },
  maxmind_asn(database, language='en'): {
    type: 'maxmind_asn',
    settings: { database: database, language: language },
  },
  maxmind_city(database, language='en'): {
    type: 'maxmind_city',
    settings: { database: database, language: language },
  },
}
