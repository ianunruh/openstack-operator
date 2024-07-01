# OpenID Connect

## Setup

In this example, the identity provider is `gitlab` and the federation protocol is `openid`. These need to be used consistently across configs and URIs for everything to work.

### Identity provider

Add a new application with a valid redirect URI.

```
https://keystone.openstack.example.com/v3/OS-FEDERATION/identity_providers/gitlab/protocols/openid/auth
```

Note the client ID and secret, create a new secret with them. A random string should be generated for `KEYSTONE_OIDC_CRYPTO_PASSPHRASE`.

```bash
kubectl create secret generic keystone-oidc \
  --from-literal=KEYSTONE_OIDC_CLIENT_ID=foo \
  --from-literal=KEYSTONE_OIDC_CLIENT_SECRET=bar \
  --from-literal=KEYSTONE_OIDC_CRYPTO_PASSPHRASE=$(pwgen 32 1)
```

### Operator

```yaml
apiVersion: openstack.ospk8s.com/v1beta1
kind: ControlPlane
metadata:
  name: default
spec:
  keystone:
    oidc:
      enabled: true
      identityProvider: gitlab
      providerMetadataURL: https://gitlab.example.com/.well-known/openid-configuration
  horizon:
    sso:
      methods:
        - kind: openid
          title: GitLab
          default: true
        - kind: credentials
          title: Keystone
```

### Keystone

The `remote-id` should match the issuer URL used in issued tokens. In this example, the `federated_users` group is being granted the `member` role on an existing project.

```bash
openstack group create federated_users
openstack role add member --group federated_users --project myproject

openstack identity provider create gitlab --remote-id https://gitlab.example.com

cat > rules.json <<EOF
[
  {
    "local": [
      {
        "user": {
          "name": "{0}"
        },
        "group": {
          "domain": {
            "name": "Default"
          },
          "name": "federated_users"
        }
      }
    ],
    "remote": [
      {
        "type": "REMOTE_USER"
      }
    ]
  }
]
EOF

openstack mapping create gitlab --rules rules.json
openstack federation protocol create openid --mapping gitlab --identity-provider gitlab
```

Federated users are assigned names based on the `sub` claim in the issued token, suffixed with the domain from the issuer URL. For example, `2@gitlab.example.com` is how the name of the federated user from a hosted GitLab instance. This can be overridden to use another claim by adding the following to the Keystone spec.

```yaml
oidc:
  extraConfig:
    OIDCRemoteUserClaim: preferred_username@
```

## References

- https://docs.openstack.org/keystone/latest/admin/federation/configure_federation.html
