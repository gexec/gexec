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
  };

  outputs = inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [
        inputs.devenv.flakeModule
      ];

      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      perSystem = { config, self', inputs', pkgs, system, ... }: {
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
              name = "genexec";

              languages = {
                go = {
                  enable = true;
                  package = pkgs.go_1_23;
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
                    "genexec"
                  ];
                };
                mysql = {
                  enable = true;
                  ensureUsers = [
                    {
                      name = "genexec";
                      password = "p455w0rd";
                      ensurePermissions = {
                        "genexec.*" = "ALL PRIVILEGES";
                      };
                    }
                  ];
                  initialDatabases = [{
                    name = "genexec";
                  }];
                };
                postgres = {
                  enable = true;
                  listen_addresses = "127.0.0.1";
                  initialScript = ''
                    CREATE USER genexec WITH ENCRYPTED PASSWORD 'p455w0rd';
                    GRANT ALL PRIVILEGES ON DATABASE genexec TO genexec;
                  '';
                  initialDatabases = [{
                    name = "genexec";
                  }];
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
                bingo
                go-task
                httpie
                nixpkgs-fmt
                sqlite
              ];

              env = {
                GENEXEC_ADMIN_USERNAME = "admin";
                GENEXEC_ADMIN_PASSWORD = "p455w0rd";
                GENEXEC_ADMIN_EMAIL = "genexec@webhippie.de";

                GENEXEC_DATABASE_DRIVER = "sqlite3";
                GENEXEC_DATABASE_NAME = "storage/genexec.sqlite3";

                GENEXEC_UPLOAD_DRIVER = "file";
                GENEXEC_UPLOAD_PATH = "storage/uploads/";
              };
            };
          };
        };
      };
    };
}
