module github.com/zot/minispec

go 1.26

require gopkg.in/yaml.v3 v3.0.1

require (
	github.com/kr/pretty v0.3.1 // indirect
	github.com/zot/simple-dom v0.0.0
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

replace github.com/zot/simple-dom => ../../mini-spec-tool
