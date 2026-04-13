# ghtoken

A tiny Go utility to generate GitHub App Installation Tokens.

## Usage
1. Name your GitHub App private key as `<APP_ID>.pem`.
2. Run:
   ```bash
   ghtoken <org_name> <path_to_key.pem>
   
ghtoken/
├── .github/
│   └── workflows/
│       └── release.yml     # The build/release automation we discussed
├── .gitignore              # Crucial: Must ignore *.pem files
├── go.mod                  # Project dependencies
├── go.sum                  # Checksums for dependencies
├── main.go                 # Your Go source code
├── README.md               # Instructions for use
└── LICENSE                 # Public repos should have an MIT/Apache license
