schema_version = 1

project {
  license        = "MPL-2.0"
  copyright_year = 2025
  copyright_holder = "Yoichiro Ito <contact.110416@gmail.com>"

  header_ignore = [
    # examples used within documentation (prose)
    "examples/**",

    # GitHub issue template configuration
    ".github/ISSUE_TEMPLATE/*.yml",

    # golangci-lint tooling configuration
    ".golangci.yml",

    # GoReleaser tooling configuration
    ".goreleaser.yml",
  ]
}
