package echoswagger

type (
	// OpenAPI represents the root object of an OpenAPI 3.1 document.
	// See https://spec.openapis.org/oas/v3.1.2.html
	OpenAPI struct {
		OpenAPI      string                `json:"openapi"`
		Info         *Info                 `json:"info"`
		Servers      []*Server             `json:"servers,omitempty"`
		Paths        map[string]*Path      `json:"paths"`
		Components   *Components           `json:"components,omitempty"`
		Security     []map[string][]string `json:"security,omitempty"`
		Tags         []*Tag                `json:"tags,omitempty"`
		ExternalDocs *ExternalDocs         `json:"externalDocs,omitempty"`
	}

	// Info provides metadata about the API. The metadata can be used by the clients if needed,
	// and can be presented in the Swagger-UI for convenience.
	Info struct {
		Title          string   `json:"title"`
		Description    string   `json:"description,omitempty"`
		TermsOfService string   `json:"termsOfService,omitempty"`
		Contact        *Contact `json:"contact,omitempty"`
		License        *License `json:"license,omitempty"`
		Version        string   `json:"version"`
	}

	// Contact contains the API contact information.
	Contact struct {
		// Name of the contact person/organization
		Name string `json:"name,omitempty"`
		// Email address of the contact person/organization
		Email string `json:"email,omitempty"`
		// URL pointing to the contact information
		URL string `json:"url,omitempty"`
	}

	// License contains the license information for the API.
	License struct {
		// Name of license used for the API
		Name string `json:"name,omitempty"`
		// URL to the license used for the API
		URL string `json:"url,omitempty"`
	}

	// Server represents a server hosting the API.
	Server struct {
		// URL to the target host. It may be relative to the location of the document.
		URL string `json:"url"`
		// Description of the host designated by the URL.
		Description string `json:"description,omitempty"`
	}

	// Components holds reusable objects referenced by the document.
	Components struct {
		Schemas         map[string]*JSONSchema     `json:"schemas,omitempty"`
		SecuritySchemes map[string]*SecurityScheme `json:"securitySchemes,omitempty"`
	}

	// Path holds the operations available on a single path.
	Path struct {
		// Ref allows for an external definition of this path item.
		Ref string `json:"$ref,omitempty"`
		// Get defines a GET operation on this path.
		Get *Operation `json:"get,omitempty"`
		// Put defines a PUT operation on this path.
		Put *Operation `json:"put,omitempty"`
		// Post defines a POST operation on this path.
		Post *Operation `json:"post,omitempty"`
		// Delete defines a DELETE operation on this path.
		Delete *Operation `json:"delete,omitempty"`
		// Options defines a OPTIONS operation on this path.
		Options *Operation `json:"options,omitempty"`
		// Head defines a HEAD operation on this path.
		Head *Operation `json:"head,omitempty"`
		// Patch defines a PATCH operation on this path.
		Patch *Operation `json:"patch,omitempty"`
		// Parameters is the list of parameters that are applicable for all the operations
		// described under this path.
		Parameters []*Parameter `json:"parameters,omitempty"`
	}

	// Operation describes a single API operation on a path.
	Operation struct {
		// Tags is a list of tags for API documentation control. Tags can be used for
		// logical grouping of operations by resources or any other qualifier.
		Tags []string `json:"tags,omitempty"`
		// Summary is a short summary of what the operation does.
		Summary string `json:"summary,omitempty"`
		// Description is a verbose explanation of the operation behavior.
		// CommonMark syntax can be used for rich text representation.
		Description string `json:"description,omitempty"`
		// ExternalDocs points to additional external documentation for this operation.
		ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"`
		// OperationID is a unique string used to identify the operation.
		OperationID string `json:"operationId,omitempty"`
		// Parameters is a list of parameters that are applicable for this operation.
		Parameters []*Parameter `json:"parameters,omitempty"`
		// RequestBody is the request body applicable for this operation.
		RequestBody *RequestBody `json:"requestBody,omitempty"`
		// Responses is the list of possible responses as they are returned from executing
		// this operation.
		Responses map[string]*Response `json:"responses,omitempty"`
		// Deprecated declares this operation to be deprecated.
		Deprecated bool `json:"deprecated,omitempty"`
		// Security is a declaration of which security schemes are applied for this operation.
		Security []map[string][]string `json:"security,omitempty"`
	}

	// Parameter describes a single operation parameter.
	Parameter struct {
		// Name of the parameter. Parameter names are case sensitive.
		Name string `json:"name"`
		// In is the location of the parameter.
		// Possible values are "query", "header" or "path".
		In string `json:"in"`
		// Description is a brief description of the parameter.
		// CommonMark syntax can be used for rich text representation.
		Description string `json:"description,omitempty"`
		// Required determines whether this parameter is mandatory.
		Required bool `json:"required"`
		// AllowEmptyValue sets the ability to pass empty-valued query parameters.
		AllowEmptyValue bool `json:"allowEmptyValue,omitempty"`
		// Schema defines the type used for the parameter.
		Schema *JSONSchema `json:"schema,omitempty"`
	}

	// RequestBody describes a single request body.
	RequestBody struct {
		// Description is a brief description of the request body.
		Description string `json:"description,omitempty"`
		// Content maps media types to their schema.
		Content map[string]*MediaType `json:"content"`
		// Required determines if the request body is required in the request.
		Required bool `json:"required,omitempty"`
	}

	// MediaType provides the schema for the media type identified by its key.
	MediaType struct {
		Schema *JSONSchema `json:"schema,omitempty"`
	}

	// Response describes an operation response.
	Response struct {
		// Description of the response. CommonMark syntax can be used for rich text representation.
		Description string `json:"description"`
		// Headers is a list of headers that are sent with the response.
		Headers map[string]*Header `json:"headers,omitempty"`
		// Content maps media types to the schema of the response.
		// If this field does not exist, no content is returned as part of the response.
		Content map[string]*MediaType `json:"content,omitempty"`
	}

	// Header represents a response header.
	Header struct {
		// Description is a brief description of the header.
		// CommonMark syntax can be used for rich text representation.
		Description string `json:"description,omitempty"`
		// Schema defines the type used for the header.
		Schema *JSONSchema `json:"schema,omitempty"`
	}

	// SecurityScheme defines a security scheme that can be used by the operations.
	// Supported schemes are HTTP authentication (basic, bearer), an API key
	// (header, query or cookie) and OAuth2's common flows.
	SecurityScheme struct {
		// Type of the security scheme. Valid values are "http", "apiKey" or "oauth2".
		Type string `json:"type"`
		// Description for security scheme
		Description string `json:"description,omitempty"`
		// Name of the header, query or cookie parameter to be used when type is "apiKey".
		Name string `json:"name,omitempty"`
		// In is the location of the API key when type is "apiKey".
		// Valid values are "query", "header" or "cookie".
		In string `json:"in,omitempty"`
		// Scheme is the HTTP authorization scheme when type is "http", e.g. "basic" or "bearer".
		Scheme string `json:"scheme,omitempty"`
		// BearerFormat is a hint to the client to identify how the bearer token is formatted.
		BearerFormat string `json:"bearerFormat,omitempty"`
		// Flows contains configuration information for the flow types supported
		// when type is "oauth2".
		Flows map[OAuth2FlowType]*OAuthFlow `json:"flows,omitempty"`
	}

	// OAuthFlow contains configuration details for a supported OAuth Flow.
	OAuthFlow struct {
		// AuthorizationURL is the authorization URL to be used for this flow.
		AuthorizationURL string `json:"authorizationUrl,omitempty"`
		// TokenURL is the token URL to be used for this flow.
		TokenURL string `json:"tokenUrl,omitempty"`
		// Scopes list the available scopes for the OAuth2 security scheme.
		Scopes map[string]string `json:"scopes"`
	}

	// ExternalDocs allows referencing an external resource for extended documentation.
	ExternalDocs struct {
		// Description is a short description of the target documentation.
		// CommonMark syntax can be used for rich text representation.
		Description string `json:"description,omitempty"`
		// URL for the target documentation.
		URL string `json:"url"`
	}

	// Tag allows adding meta data to a single tag that is used by the Operation Object. It is
	// not mandatory to have a Tag Object per tag used there.
	Tag struct {
		// Name of the tag.
		Name string `json:"name,omitempty"`
		// Description is a short description of the tag.
		// CommonMark syntax can be used for rich text representation.
		Description string `json:"description,omitempty"`
		// ExternalDocs is additional external documentation for this tag.
		ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"`
	}

	// JSONSchema is the subset of JSON Schema 2020-12 generated by echoswagger.
	JSONSchema struct {
		Ref                  string                 `json:"$ref,omitempty"`
		Type                 JSONType               `json:"type,omitempty"`
		Format               string                 `json:"format,omitempty"`
		Description          string                 `json:"description,omitempty"`
		Items                *JSONSchema            `json:"items,omitempty"`
		Properties           map[string]*JSONSchema `json:"properties,omitempty"`
		AdditionalProperties *JSONSchema            `json:"additionalProperties,omitempty"`
		Required             []string               `json:"required,omitempty"`
		DefaultValue         interface{}            `json:"default,omitempty"`
		Examples             []interface{}          `json:"examples,omitempty"`
		Enum                 []interface{}          `json:"enum,omitempty"`
		Minimum              *float64               `json:"minimum,omitempty"`
		Maximum              *float64               `json:"maximum,omitempty"`
		MinLength            *int                   `json:"minLength,omitempty"`
		MaxLength            *int                   `json:"maxLength,omitempty"`
		ReadOnly             bool                   `json:"readOnly,omitempty"`
		ContentMediaType     string                 `json:"contentMediaType,omitempty"`
		XML                  *XMLSchema             `json:"xml,omitempty"`
	}

	// JSONType is the JSON type enum.
	JSONType string

	// XMLSchema adds metadata for the XML representation of a property.
	XMLSchema struct {
		Name      string `json:"name,omitempty"`
		Namespace string `json:"namespace,omitempty"`
		Prefix    string `json:"prefix,omitempty"`
		Attribute bool   `json:"attribute,omitempty"`
		Wrapped   bool   `json:"wrapped,omitempty"`
	}
)
