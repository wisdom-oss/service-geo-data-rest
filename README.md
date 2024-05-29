<div align="center">
<img height="150px" src="https://raw.githubusercontent.com/wisdom-oss/brand/main/svg/standalone_color.svg">

<!-- TODO: Change Information here -->

<h1>Geospatial Data Service</h1>
<h3>geodata-service</h3>
<p>🌍🗺️ a service handling access to shapes and other data</p>

<!-- TODO: Change URL here to point to correct repository -->
<img src="https://img.shields.io/github/go-mod/go-version/wisdom-oss/service-geo-data-restd?style=for-the-badge" alt="Go Lang Version"/>
<a href="openapi.yaml">
<img src="https://img.shields.io/badge/Schema%20Version-3.0.0-6BA539?style=for-the-badge&logo=OpenAPI%20Initiative" alt="Open
API Schema Version"/></a>
</div>

This microservice handles the access to predefined layers and returns them in
a GeoJSON response.
It currently only supports returning a full layer or selection of it by 
attributes and not geospatial relations.
Furthermore, uploading a new shape is not supported at this point in time
