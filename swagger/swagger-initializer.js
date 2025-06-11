window.onload = function() {
  window.ui = SwaggerUIBundle({
          url: "app:8000/swagger/doc.json",
    dom_id: '#swagger-ui',
    schemes: ["http"],  // ← Force HTTP here
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ]
  });
};