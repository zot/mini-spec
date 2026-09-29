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

Project -> os: Stat(rootPath + "/.minispec.toml")
alt config exists
    Project -> toml: Decode(file)
    toml --> Project: Config, undecoded keys
else no config
    Project -> Project: use defaults
end

Project --> CLI: Project{rootPath, designDir, srcDir, config}
```

# Default Configuration

```toml
design_dir = "design"
src_dir = "src"
code_extensions = [".go", ".ts", ".js", ".lua", ".py", ".c", ".h", ".cpp", ".sh"]
```
