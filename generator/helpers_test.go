package generator_test

import (
	"testing"

	"github.com/kalbasit/sqlc-multi-db/generator"
)

func TestFixAcronyms(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple Id to ID",
			input:    "userId",
			expected: "userID",
		},
		{
			name:     "Url to URL",
			input:    "profileUrl",
			expected: "profileURL",
		},
		{
			name:     "Api to API",
			input:    "fetchApiData",
			expected: "fetchAPIData",
		},
		{
			name:     "Sql to SQL",
			input:    "execSqlQuery",
			expected: "execSQLQuery",
		},
		{
			name:     "Json to JSON",
			input:    "parseJsonBody",
			expected: "parseJSONBody",
		},
		{
			name:     "Identifier should not be corrupted",
			input:    "Identifier",
			expected: "Identifier",
		},
		{
			name:     "Curling should not be corrupted to CURLing",
			input:    "Curling",
			expected: "Curling",
		},
		{
			name:     "XmlParser should become XMLParser",
			input:    "XmlParser",
			expected: "XMLParser",
		},
		{
			name:     "HtmlDocument should become HTMLDocument",
			input:    "HtmlDocument",
			expected: "HTMLDocument",
		},
		{
			name:     "Multiple acronyms in one string",
			input:    "userId and profileUrl",
			expected: "userID and profileURL",
		},
		{
			name:     "Id at end of string",
			input:    "GetUserId",
			expected: "GetUserID",
		},
		{
			name:     "Already correct ID should stay",
			input:    "GetUserID",
			expected: "GetUserID",
		},
		{
			name:     "Already correct URL should stay",
			input:    "profileURL",
			expected: "profileURL",
		},
		{
			name:     "Tcp connection",
			input:    "openTcpConnection",
			expected: "openTCPConnection",
		},
		{
			name:     "Jwt token",
			input:    "validateJwtToken",
			expected: "validateJWTToken",
		},
		{
			name:     "Ec2 instance",
			input:    "launchEc2Instance",
			expected: "launchEC2Instance",
		},
		{
			name:     "Json web token Jwt",
			input:    "parseJsonWebToken",
			expected: "parseJSONWebToken",
		},
		{
			name:     "Uuid should not change (not in our list)",
			input:    "userUuid",
			expected: "userUuid",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "LastInsertId in method call should not be transformed",
			input:    "res.LastInsertId()",
			expected: "res.LastInsertId()",
		},
		{
			name:     "SetId in method call should not be transformed",
			input:    "obj.SetId(123)",
			expected: "obj.SetId(123)",
		},
		{
			name:     "userId in variable assignment should become userID",
			input:    "userId = 123",
			expected: "userID = 123",
		},
		{
			name:     "NewUrl in struct literal should remain NewUrl",
			input:    "NewUrl: \"value\"",
			expected: "NewUrl: \"value\"",
		},
		{
			name:     "OldUrl in struct literal should remain OldUrl",
			input:    "OldUrl: \"value\"",
			expected: "OldUrl: \"value\"",
		},
		{
			name:     "profileUrl in assignment should become profileURL",
			input:    "profileUrl = \"https://example.com\"",
			expected: "profileURL = \"https://example.com\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := generator.FixAcronyms([]byte(tt.input))
			if string(result) != tt.expected {
				t.Errorf("FixAcronyms(%q) = %q, want %q", tt.input, string(result), tt.expected)
			}
		})
	}
}
