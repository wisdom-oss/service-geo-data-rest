-- name: get-layers
SELECT *
FROM geodata.layers;

-- name: get-layer
SELECT *
FROM geodata.layers
WHERE id = $1;

-- name: get-layer-contents
SELECT *
FROM geodata.%s ;
