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

                packages = with pkgs; [
                  cosign
                  go-task
                  httpie
                  nixfmt-rfc-style
                  sqlite
                  yq-go
                ];

                env = {
                  GEXEC_TOKEN_SECRET = "NTaCR5JztYujaOZNgesaUzaVPmoxkGo0";

                  GEXEC_ADMIN_USERNAME = "admin";
                  GEXEC_ADMIN_PASSWORD = "p455w0rd";
                  GEXEC_ADMIN_EMAIL = "gexec@webhippie.de";

                  GEXEC_DATABASE_DRIVER = "sqlite3";
                  GEXEC_DATABASE_NAME = "storage/gexec.sqlite3";

                  GEXEC_UPLOAD_DRIVER = "file";
                  GEXEC_UPLOAD_PATH = "storage/uploads/";
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
                  redis = {
                    enable = true;
                  };
                };

                processes = {
                  gexec-server = {
                    exec = "task watch:server";

                    process-compose = {
                      readiness_probe = {
                        exec.command = "${pkgs.curl}/bin/curl -sSf http://localhost:8000/readyz";
                        initial_delay_seconds = 2;
                        period_seconds = 10;
                        timeout_seconds = 4;
                        success_threshold = 1;
                        failure_threshold = 5;
                      };

                      availability = {
                        restart = "on_failure";
                      };
                    };
                  };

                  gexec-worker = {
                    exec = "task watch:runner";

                    process-compose = {
                      readiness_probe = {
                        exec.command = "${pkgs.curl}/bin/curl -sSf http://localhost:8001/readyz";
                        initial_delay_seconds = 2;
                        period_seconds = 10;
                        timeout_seconds = 4;
                        success_threshold = 1;
                        failure_threshold = 5;
                      };

                      availability = {
                        restart = "on_failure";
                      };
                    };
                  };

                  gexec-webui = {
                    exec = "task watch:frontend";

                    process-compose = {
                      readiness_probe = {
                        exec.command = "${pkgs.curl}/bin/curl -sSf http://localhost:5173";
                        initial_delay_seconds = 2;
                        period_seconds = 10;
                        timeout_seconds = 4;
                        success_threshold = 1;
                        failure_threshold = 5;
                      };

                      availability = {
                        restart = "on_failure";
                      };
                    };
                  };

                  minio = {
                    process-compose = {
                      readiness_probe = {
                        exec.command = "${pkgs.curl}/bin/curl -sSf http://localhost:9000/minio/health/live";
                        initial_delay_seconds = 2;
                        period_seconds = 10;
                        timeout_seconds = 4;
                        success_threshold = 1;
                        failure_threshold = 5;
                      };

                      availability = {
                        restart = "on_failure";
                      };
                    };
                  };
                };
              };
            };
          };
        };
    };
}
