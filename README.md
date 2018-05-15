# Genderize

[![Github license](https://img.shields.io/github/license/SteelPangolin/go-genderize.svg?style=flat)](https://github.com/SteelPangolin/go-genderize)
[![Travis CI build status](https://img.shields.io/travis/SteelPangolin/go-genderize.svg?style=flat)](https://travis-ci.org/SteelPangolin/go-genderize)
[![Codecov code coverage](https://img.shields.io/codecov/c/github/SteelPangolin/go-genderize.svg?style=flat)](https://codecov.io/gh/SteelPangolin/go-genderize)

Go client for the [Genderize.io](https://genderize.io/) web service.

[Full API documentation is available on GoDocs](https://godoc.org/github.com/SteelPangolin/go-genderize).

## Basic usage

Simple interface with minimal configuration.

### Code

```go
responses, err := Get([]string{"James", "Eva", "Thunderhorse"})
if err != nil {
    panic(err)
}
for _, response := range responses {
    fmt.Printf("%s: %s\n", response.Name, response.Gender)
}
```

### Output

```
James: male
Eva: female
Thunderhorse:
```

## Advanced usage

Client with custom API key and user agent, query with language and country IDs.

### Code

```go
client, err := NewClient(Config{
    UserAgent: "GoGenderizeDocs/0.0",
    // Note that you'll need to use your own API key.
    APIKey: "",
})
if err != nil {
    panic(err)
}
responses, err := client.Get(Query{
    Names:      []string{"Kim"},
    CountryID:  "dk",
    LanguageID: "da",
})
if err != nil {
    panic(err)
}
for _, response := range responses {
    fmt.Printf("%s: %s\n", response.Name, response.Gender)
}
```

### Output

```
Kim: male
```

## Release checklist

1. Generate a new version number: `major.minor.micro`. It should be compatible with [SemVer 2.0.0](https://semver.org/).
2. Update `Version` in `genderize.go`.
3. Add a changelog entry and date for the new version in `CHANGES.md`.
4. Commit the changes. This may be done as part of another change.
5. Tag the commit with `git tag major.minor.micro`.
6. Push the tag to GitHub with `git push origin major.minor.micro`.
