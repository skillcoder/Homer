# install /usr/ports/databases/timescaledb
# Add timescaledb to shared_preload_libraries in /var/db/postgres/data10/postgresql.conf
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;
# Or if upgrade ALTER EXTENSION timescaledb UPDATE TO '%%PORTVERSION%%';
