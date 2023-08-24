# Kubernetes Image Warden

<p align="center">
<img src="./docs/logo.png" alt="logo" width="300"/>
</p>

Kubernetes Image Warden (**KIW**) is an open-source application designed to provide comprehensive image management and enforce pull policies in Kubernetes clusters. 
It acts as a vigilant guardian, ensuring the security, compliance, and efficient usage of container images within your Kubernetes environment.

## Project Status: Experimental

Kubernetes Image Warden is currently in an experimental phase, undergoing active development and testing.
Users should be aware of possible bugs and limitations while providing feedback to shape its future progress. 
It is not recommended for production use at this stage

## Example Use Cases

Discover how KIW simplifies image management in your Kubernetes clusters with these powerful use cases.

#### Rolling Tag Management

Easily manage rolling tags (e.g., `latest` or `v1`) for container images, ensuring consistent versioning across all nodes by managing blocks and allowing lists for particular images with rolling tags. KIW can automatically detect images with rolling tags by keeping track of historical images present for each node.

### Distributed Image Monitoring

Effortlessly track container image presence across your distributed Kubernetes setup. 
KIW provides `docker images` command experience that works across distributed Kubernetes nodes.

See more detailed [usage](./docs/usage.md) documentation.

## Alternatives

- [Kyverno](https://kyverno.io) is a powerful Kubernetes policy engine primarily focused on enforcing policies for resources like pods, deployments, and namespaces. While Kyverno is excellent for resource policy enforcement, Kubernetes Image Warden (KIW) offers an advantage with its ability to keep a historical track of images. This feature enables KIW to provide more comprehensive and dynamic image management policies.

- [crictl](https://github.com/kubernetes-sigs/cri-tools/tree/master) is a command-line utility that allows direct interaction with container runtimes on Kubernetes nodes. While crictl offers low-level control over container operations, KIW's distributed image monitoring makes it more suitable for multi-node Kubernetes environments. KIW effortlessly tracks image activities across distributed nodes, providing real-time visibility into image pulls and usage patterns.

In comparison, Kubernetes Image Warden stands out as a specialized tool for image management, ensuring better image policies with historical tracking and seamless usage in distributed Kubernetes setups.

## Key Components

KIW is composed of three core components, each playing a crucial role in maintaining the integrity and control of container images:

1. **Agent**: The Image Warden agent is a lightweight and efficient component that runs on each node within the Kubernetes cluster. 
It continuously monitors the containers running on the node, tracks image usage, and reports relevant data to the central controller.

2. **Controller**: The Controller is the heart of KIW, responsible for managing rules and data collection from the agents. 
It acts as the centralized brain that enforces image policies, conducts image analysis, and maintains a repository of image-related information.

3. **CLI (Command-Line Interface)**: To facilitate seamless interaction with the Kubernetes Image Warden, we provide a user-friendly command-line interface. 
The CLI empowers users to perform basic operations, configure policies, and gain insights into image usage within the Kubernetes cluster.

## Key Features

KIW boasts a wide array of powerful features, enabling you to optimize, secure, and streamline your Kubernetes image management:

* **Pull Policies Enforcement**: Define and enforce pull policies for container images to ensure compliance and avoid unauthorized access.

* **Image Analysis and Optimization**: Gain valuable insights into image sizes, usage patterns, and metadata to optimize resource allocation and streamline your container deployments.

* **Real-time Monitoring**: Monitor image-related activities in real-time, allowing you to detect and respond promptly to potential security threats.

* **Customizable Rules**: Tailor the behavior of Kubernetes Image Warden to align with your organization's specific requirements through flexible and customizable rules.

## Getting Started

To start using Kubernetes Image Warden, follow the step-by-step guide in our [Installation guideline](./docs/installation.md). 
Whether you are an administrator seeking enhanced image security or a developer aiming to optimize image usage, KIW has you covered.

## Contributing guideline

We welcome contributions from the community to improve Kubernetes Image Warden. To get started, follow these simple steps:

1. **Check for Existing Issues**: Look through the issue tracker to find any open bugs or feature requests you'd like to work on. 
If you don't find one, you can propose new features by opening a new issue.

2. **Fork the Repository**: Fork the official repository to your GitHub account.

3. **Create a Branch**: Make a new branch in your fork to work on your changes. Give it a descriptive name related to your contribution.

4. **Make Changes**: Write your code and make the necessary improvements or fixes.

5. **Test**: Ensure your changes don't break anything and add tests if applicable.

6. **Submit a Pull Request**: Once you're ready, submit a pull request from your branch to the main repository's appropriate branch.

7. **Review**: The maintainers will review your pull request, provide feedback, and collaborate with you to improve it.

8. **CLAs (if applicable)**: If required, sign the Contributor License Agreement.

9. **Merge and Acknowledgment**: Once your pull request is accepted and merged, your contribution will be acknowledged in the project.

## License

Kubernetes Image Warden is released under the [MIT License](./LICENSE), making it free to use, modify, and distribute. 
Please review the license file for more details.