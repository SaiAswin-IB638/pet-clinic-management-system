FROM nginx:alpine

# Copy Nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy prebuilt Swagger UI
COPY swagger-ui /usr/share/nginx/html/swagger-ui

# Fix Swagger UI config to point to /swagger.json
RUN sed -i 's|https://petstore.swagger.io/v2/swagger.json|/swagger.json|g' \
    /usr/share/nginx/html/swagger-ui/swagger-initializer.js