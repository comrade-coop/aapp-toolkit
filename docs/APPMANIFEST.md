# :rocket: aApp Toolkit: Application Manifest Documentation

Welcome to the **AAP-Toolkit** documentation! This guide explains how to define and configure your application using the **app-manifest** (in JSON). The `app-manifest` helps you provision a **Confidential VM** and start your containerized app with the necessary volumes, environment variables, and ingress settings.

Below, you’ll find **snippets** of a simplified `app-manifest`, best practices, and explanations for the key fields.

---

## :eyes: Overview

- **Confidential VM Provisioning**: Launches a secure VM for your app.  
- **Container Configuration**: Defines how to build and run your container image.  
- **Volumes**:  
  - `cloud` volume: Shared volume that can be transferred between instances after mutual attestation.  
  - `local` volume: Regular storage that persists only on the local instance.  
- **Ingress**: Configures external access to your app.

---

## :page_facing_up: Example Manifest Snippets

> **Note**: The snippets below demonstrate the **essential** fields of an `app-manifest`. Omitted fields (marked as `...`) can be included as needed.

### :pushpin: Metadata & Basic Structure

```json
{
  "apiVersion": "aapp-toolkit.comrade.coop/v1alpha1",
  "kind": "Application",
  "metadata": {
    "annotations": {
      "repo": "https://github.com/example-user/aapp-repo.git",
      "tag": "v1.0.0"
    }
  },
  "spec": {
    ...
  }
}
```

- **`apiVersion`**: Identifies the AAP-Toolkit API version.  
- **`kind`**: Resource type, here always `"Application"`.  
- **`metadata.annotations.repo`**: Git repo URL of your application.  
- **`metadata.annotations.tag`**: Tag or version reference.

### :wrench: Container Build & Volumes

```json
{
  "spec": {
    "container": {
      "build": {
        "repo": "https://github.com/example-user/my-app.git",
        "tag": "my-app:latest",
        "dockerfile": "Dockerfile",
        "args": {
          "BACKEND_ETH_RPC": "https://cloud.example.com/rpc"
        }
      },
      "volumes": [
        {
          "name": "secrets-vol",
          "mount": "/shared_secrets",
          "type": "cloud"
        },
        {
          "name": "data-vol",
          "mount": "/data",
          "type": "local"
        }
      ]
    },
    ...
  }
}
```

- **`build.repo`**: Git repo for building your container image.  
- **`build.tag`**: Container image tag (e.g., `my-app:latest`).  
- **`build.args`**: Build-time arguments, e.g., environment variables.  
- **`volumes`**: Two volume types:
  - **Cloud volume** (`"type": "cloud"`): Securely shared between app instances.
  - **Local volume** (`"type": "local"`): Standard storage, local to the container.

### :globe_with_meridians: Ingress Configuration & Developer Key

```json
{
  "spec": {
    ...
    "ingress": {
      "hostnames": ["app.example.com", "console.app.example.com"],
      "port": 80
    },
    "developer": {
      "key": "0xEXAMPLEDEVPUBKEY"
    }
  }
}
```

- **`ingress.hostnames`**: Domains for external access.  
- **`ingress.port`**: Port through which the app is exposed.  
- **`developer.key`**: Developer’s public key used for authentication.

---

## :card_file_box: Volume Types & Mutual Attestation

1. **Local Volume**  
   - Standard container storage.  
   - Lives and dies with the instance.  
   - Example usage: Temporary files, caches.  

2. **Cloud Volume**  
   - Persistent storage shared across instances.  
   - If a **new** instance is created, it performs **mutual attestation** with an existing instance.  
   - Once attested, the old instance transfers the content of the shared volume to the new instance.  
   - If the cloud volume is empty (e.g., on the first deploy), the app generates secrets; otherwise, it reuses existing secrets.

---

## :lock: Security & Confidential VM

- **Confidential VM**: Provides enhanced security by ensuring the VM is tamper-resistant.  
- **Mutual Attestation**: Verifies both the old and new instance before exchanging sensitive data (like `cloud` volume contents).

---

## :memo: Conclusion

By defining your **app-manifest** correctly, you enable:

- Secure, scalable deployment with **Confidential VM** instances.  
- Seamless volume sharing across instances with **mutual attestation**.  
- Flexible ingress and build configuration for easy maintenance.

Use the snippets in this guide to craft your own minimal `app-manifest`. For more advanced features, refer to the extended documentation and examples in the repository.  
