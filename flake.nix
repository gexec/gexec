{
  description = "Nix flake for development";

  inputs = {
    nixpkgs = {
      url = "github:nixos/nixpkgs/nixpkgs-unstable";
    };

    devenv = {
      url = "github:cachix/devenv";
    };

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
    };

    git-hooks = {
      url = "github:cachix/git-hooks.nix";
    };
  };

  outputs =
    inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [
        inputs.devenv.flakeModule
        inputs.git-hooks.flakeModule
      ];

      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      perSystem =
        {
          config,
          self',
          inputs',
          pkgs,
          system,
          ...
        }:
        {
          imports = [
            {
              _module.args.pkgs = import inputs.nixpkgs {
                inherit system;
                config.allowUnfree = true;
              };
            }
          ];

          devenv = {
            shells = {
              default = {
                name = "gexec";

                git-hooks = {
                  hooks = {
                    nixfmt-rfc-style = {
                      enable = true;
                    };

                    gofmt = {
                      enable = true;
                    };

                    golangci-lint = {
                      enable = true;
                    };
                  };
                };

                languages = {
                  go = {
                    enable = true;
                    package = pkgs.go_1_24;
                  };
                  javascript = {
                    enable = true;
                    package = pkgs.nodejs_20;
                  };
                };

                services = {
                  minio = {
                    enable = true;
                    accessKey = "WMHJQCKGKWUW1CRGPQ8Y";
                    secretKey = "CKchNYh6D1mn1Vs6XMfnDmuK76PZ3XE3vF56LDS0";
                    buckets = [
                      "gexec"
                    ];
                  };
                  mysql = {
                    enable = true;
                    ensureUsers = [
                      {
                        name = "gexec";
                        password = "p455w0rd";
                        ensurePermissions = {
                          "gexec.*" = "ALL PRIVILEGES";
                        };
                      }
                    ];
                    initialDatabases = [
                      {
                        name = "gexec";
                      }
                    ];
                  };
                  postgres = {
                    enable = true;
                    listen_addresses = "127.0.0.1";
                    initialScript = ''
                      CREATE USER gexec WITH ENCRYPTED PASSWORD 'p455w0rd';
                      GRANT ALL PRIVILEGES ON DATABASE gexec TO gexec;
                    '';
                    initialDatabases = [
                      {
                        name = "gexec";
                      }
                    ];
                  };
                };

                processes = {
                  server = {
                    exec = "task watch:server";
                  };

                  runner = {
                    exec = "task watch:runner";
                  };
                };

                packages = with pkgs; [
                  cosign
                  go-task
                  httpie
                  nixfmt-rfc-style
                  sqlite
                  yq
                ];

                env = {
                  GENEXEC_TOKEN_SECRET = "NTaCR5JztYujaOZNgesaUzaVPmoxkGo0";

                  GENEXEC_ADMIN_USERNAME = "admin";
                  GENEXEC_ADMIN_PASSWORD = "p455w0rd";
                  GENEXEC_ADMIN_EMAIL = "gexec@webhippie.de";

                  GENEXEC_DATABASE_DRIVER = "sqlite3";
                  GENEXEC_DATABASE_NAME = "storage/gexec.sqlite3";

                  GENEXEC_UPLOAD_DRIVER = "file";
                  GENEXEC_UPLOAD_PATH = "storage/uploads/";
                };
              };
            };
          };
        };
    };
}
