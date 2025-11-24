# Kyma Serverless Functions - Documentation for LLMs

## Overview

Kyma Serverless Functions are Kubernetes custom resources that enable serverless computing on Kyma clusters. Functions are simple code snippets that implement specific business logic without requiring server management. They are event-driven, automatically scaled, and integrate seamlessly with the Kyma ecosystem.

## Key Concepts

- **Function**: A Kubernetes custom resource (CRD) that defines code, runtime, and configuration
- **Event-Driven**: Functions are triggered by events or API calls, not continuously running
- **Automatic Scaling**: Scale to zero when idle, scale up based on demand
- **Runtime Isolation**: Each function runs in its own Pod with dedicated resources

## API Group and Version

- **API Group**: `serverless.kyma-project.io`
- **API Version**: `v1alpha2`
- **Kind**: `Function`
- **Plural**: `functions`
- **Scope**: Namespaced

## Supported Runtimes

Functions support multiple programming language runtimes:

- **`nodejs20`**: Node.js v20 LTS
- **`nodejs22`**: Node.js v22 LTS
- **`python312`**: Python 3.12

## Function Spec Structure

### Required Fields

```yaml
apiVersion: serverless.kyma-project.io/v1alpha2
kind: Function
metadata:
  name: my-function
  namespace: default
spec:
  runtime: nodejs20  # Required: nodejs20, nodejs22, or python312
  source:            # Required: Defines code source
    inline:          # Option 1: Inline code
      source: |
        module.exports = {
          main: async function (event, context) {
            return "Hello World!";
          }
        }
```

### Source Options

#### 1. Inline Source

Provide code directly in the Function resource:

```yaml
spec:
  source:
    inline:
      source: |
        // Your code here
      dependencies: |
        {
          "dependencies": {
            "axios": "^1.0.0"
          }
        }
```

#### 2. Git Repository Source

Reference code from a Git repository:

```yaml
spec:
  source:
    gitRepository:
      url: "https://github.com/user/repo"
      baseDir: "/functions"              # Optional: subdirectory
      reference: "main"                    # Branch, tag, or commit
      auth:                                # Optional: for private repos
        type: basic                        # or 'key' for SSH
        secretName: git-credentials
```

### Resource Configuration

#### Resource Profiles

Pre-defined resource allocations:

```yaml
spec:
  resourceProfile: L  # XS, S, M, L, XL
```

#### Custom Resources

Define exact CPU and memory limits:

```yaml
spec:
  resources:
    limits:
      cpu: "200m"
      memory: "256Mi"
    requests:
      cpu: "100m"
      memory: "128Mi"
```

### Scaling Configuration

Control replica counts and auto-scaling:

```yaml
spec:
  replicas:
    min: 1          # Minimum replicas (0 for scale-to-zero)
    max: 10         # Maximum replicas
```

### Environment Variables

Inject configuration and secrets:

```yaml
spec:
  env:
    - name: API_KEY
      value: "my-api-key"
    - name: DB_PASSWORD
      valueFrom:
        secretKeyRef:
          name: db-secret
          key: password
```

### Labels and Annotations

Apply custom metadata to Function Pods:

```yaml
spec:
  labels:
    team: backend
    app: my-app
  annotations:
    description: "Processes user events"
```

## Complete Example: Node.js Function with Git Source

```yaml
apiVersion: serverless.kyma-project.io/v1alpha2
kind: Function
metadata:
  name: event-processor
  namespace: production
spec:
  runtime: nodejs22
  source:
    gitRepository:
      url: "https://github.com/myorg/functions"
      baseDir: "/event-processor"
      reference: "v1.2.0"
  resourceProfile: M
  replicas:
    min: 2
    max: 20
  env:
    - name: LOG_LEVEL
      value: "info"
    - name: API_ENDPOINT
      valueFrom:
        configMapKeyRef:
          name: api-config
          key: endpoint
  labels:
    app: event-system
    version: v1.2.0
```

## Complete Example: Python Function with Inline Source

```yaml
apiVersion: serverless.kyma-project.io/v1alpha2
kind: Function
metadata:
  name: data-transformer
  namespace: default
spec:
  runtime: python312
  source:
    inline:
      source: |
        def main(event, context):
            data = event.get('data', {})
            # Transform data
            result = {
                'status': 'success',
                'transformed': data
            }
            return result
      dependencies: |
        requests==2.31.0
        pandas==2.1.0
  resourceProfile: S
  replicas:
    min: 0  # Scale to zero when idle
    max: 5
```

## Function Lifecycle and Status

Functions progress through several states:

1. **ConfigurationReady**: Function configuration is validated
2. **BuildReady**: Function code is built into a container image
3. **Running**: Function is deployed and ready to receive requests

Check status with:
```bash
kubectl get functions -n namespace
```

## Best Practices for Creating Functions

1. **Choose the Right Runtime**: Match the runtime to your team's expertise
2. **Start with Resource Profile**: Use predefined profiles before custom resources
3. **Enable Scale-to-Zero**: Set `replicas.min: 0` for cost efficiency on low-traffic functions
4. **Use Git Sources for Production**: Inline is good for dev/test, Git for production
5. **Inject Secrets Securely**: Never hardcode secrets, use `valueFrom.secretKeyRef`
6. **Add Meaningful Labels**: Helps with organization and monitoring
7. **Test Locally**: Kyma provides tools for local testing before deployment

## Common Patterns

### HTTP API Function (Node.js)

```yaml
spec:
  runtime: nodejs20
  source:
    inline:
      source: |
        module.exports = {
          main: async function (event, context) {
            const { method, path, body } = event;

            if (method === 'GET' && path === '/health') {
              return { statusCode: 200, body: 'OK' };
            }

            return {
              statusCode: 200,
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ message: 'Processed' })
            };
          }
        }
```

### Event Handler Function (Python)

```yaml
spec:
  runtime: python312
  source:
    inline:
      source: |
        import json

        def main(event, context):
            event_type = event.get('type')
            event_data = event.get('data')

            if event_type == 'order.created':
                # Process order
                return {'status': 'order_processed'}

            return {'status': 'event_ignored'}
```

## Troubleshooting

### Function Not Starting

- Check `kubectl describe function <name>`
- Verify runtime is valid (nodejs20, nodejs22, python312)
- Ensure source code is syntactically correct
- Check resource limits are not too restrictive

### Build Failures

- Validate dependencies syntax
- Check Git repository is accessible
- Verify branch/tag reference exists
- Review build logs: `kubectl logs -l app=<function-name>`

### Performance Issues

- Increase resource profile or custom resources
- Adjust `replicas.max` for higher scale
- Review function code for inefficiencies
- Check external API latencies

## Integration with MCP

When using generated MCP toolsets:

- **create**: Deploy a new Function
- **get**: Retrieve Function details and status
- **list**: List all Functions in a namespace
- **update**: Modify Function configuration or code
- **delete**: Remove a Function

The generated toolset handles JSON schema validation and Kubernetes API calls automatically.
