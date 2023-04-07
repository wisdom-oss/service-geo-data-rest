-- name: create-schema
create schema if not exists geodata;

-- name: create-table
create table if not exists geodata.shapes (
    id integer not null default nextval('shapes_id_seq'::regclass),
    geom public.geometry(MultiPolygon),
    key character varying(12),
    name character varying(254),
    nuts_key character varying(254)
);

-- name: get-all-shapes
SELECT name, key, nuts_key, ST_AsGeoJSON(geom)
FROM geodata.shapes;

-- name: get-box-for-all-shapes
SELECT ST_ASGeoJson(ST_FlipCoordinates(ST_Extent(geom)))
FROM geodata.shapes;

-- name: get-shapes-by-resolution
SELECT name, key, nuts_key, ST_ASGeoJSON(geom)
FROM geodata.shapes
WHERE length(key) = $1;

-- name: get-box-for-shapes-by-resolution
SELECT ST_ASGeoJson(ST_FlipCoordinates(ST_Extent(geom)))
FROM geodata.shapes
WHERE length(key) = $1;

-- name: get-shapes-by-key
SELECT name, key, nuts_key, ST_ASGeoJSON(geom)
FROM geodata.shapes
WHERE key = any($1);

--name: get-box-for-shapes-by-key
SELECT ST_ASGeoJson(ST_FlipCoordinates(ST_Extent(geom)))
FROM geodata.shapes
WHERE key = any($1);

-- name: get-shapes-by-key-resolution
SELECT name, key, nuts_key, ST_ASGeoJSON(geom)
FROM geodata.shapes
WHERE length(key) = $1
AND key ~ $2;

-- name: get-box-for-shapes-by-key-resolution
SELECT ST_ASGeoJson(ST_FlipCoordinates(ST_Extent(geom)))
FROM geodata.shapes
WHERE length(key) = $1
AND key ~ $2;