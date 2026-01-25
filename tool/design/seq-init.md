# Sequence: Project Detection

```
User -> CLI: minispec <command>
CLI -> Project: Detect()

Project -> os: Getwd()
os --> Project: /home/user/myproject/src

loop walk up directories
    Project -> os: Stat(dir + "/design")
    alt design/ exists
        Project -> Project: rootPath = dir
        break
    else not found
        Project -> Project: dir = parent(dir)
    end
end

alt no design/ found
    Project --> CLI: error "no design/ directory found"
    CLI --> User: exit 1
end

Project -> os: Stat(rootPath + "/.minispec.yaml")
alt config exists
    Project -> yaml: Unmarshal(file)
    yaml --> Project: Config
else no config
    Project -> Project: use defaults
end

Project --> CLI: Project{rootPath, designDir, srcDir, config}
```

# Default Configuration

```yaml
design_dir: design
src_dir: src
code_extensions: [.go, .ts, .js, .lua, .py]
traceability_pattern: "// CRC:"
```
