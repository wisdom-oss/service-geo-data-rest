-- name: get-layers
SELECT *
FROM geodata.layers;

-- name: get-layer
SELECT *
FROM geodata.layers
WHERE id = $1;

-- name: get-layer-contents
SELECT id, st_transform(geometry, 4326) as geometry, key, name, additional_properties
FROM geodata.%s ;
