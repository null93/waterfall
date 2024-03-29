# Waterfall
> Cloudformation CLI tool to analyze and visualize stack events as a waterfall diagram

![GitHub Release](https://img.shields.io/github/v/release/null93/waterfall?sort=semver&style=for-the-badge&color=%23007EC6)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/null93/waterfall/.github%2Fworkflows%2Fmain.yml?style=for-the-badge&color=%23007EC6)

<p align="center" >
  <img width="75%" alt="screenshot" src="https://github.com/null93/waterfall/assets/5500199/4e01d4ef-0abf-4701-8350-ef4c4e23cbfe">
</p>

## About

In the realm of cloud infrastructure management, the default CloudFormation dashboard often lacks depth, offering only a basic list of events. This can make it challenging to gain a comprehensive understanding of the lifecycle of resources. This program seeks to overcome this limitation by gathering and organizing events from your primary stack and its nested counterparts. The outcome is a thorough interval-based visualization presented in an intuitive waterfall diagram. The status of each interval (in-progress, complete, or failed) is represented using different Unicode characters, while the type of event is indicated by color (green for create, red for delete, etc).

## Install

<details>
  <summary>Darwin</summary>

  ### Intel & ARM
  
  ```shell
  brew tap null93/tap
  brew install waterfall
  ```
</details>

<details>
  <summary>Debian</summary>

  ### amd64
  
  ```shell
  curl -sL -o ./waterfall_1.0.1_amd64.deb https://github.com/null93/waterfall/releases/download/1.0.1/waterfall_1.0.1_amd64.deb
  sudo dpkg -i ./waterfall_1.0.1_amd64.deb
  rm ./waterfall_1.0.1_amd64.deb
  ```

  ### arm64

  ```shell
  curl -sL -o ./waterfall_1.0.1_arm64.deb https://github.com/null93/waterfall/releases/download/1.0.1/waterfall_1.0.1_arm64.deb
  sudo dpkg -i ./waterfall_1.0.1_arm64.deb
  rm ./waterfall_1.0.1_arm64.deb
  ```
</details>

<details>
  <summary>Red Hat</summary>
  
  ### aarch64

  ```shell
  rpm -i https://github.com/null93/waterfall/releases/download/1.0.1/waterfall-1.0.1-1.aarch64.rpm
  ```

  ### x86_64

  ```shell
  rpm -i https://github.com/null93/waterfall/releases/download/1.0.1/waterfall-1.0.1-1.x86_64.rpm
  ```
</details>

<details>
  <summary>Alpine</summary>
  
  ### aarch64

  ```shell
  curl -sL -o ./waterfall_1.0.1_aarch64.apk https://github.com/null93/waterfall/releases/download/1.0.1/waterfall_1.0.1_aarch64.apk
  apk add --allow-untrusted ./waterfall_1.0.1_aarch64.apk
  rm ./waterfall_1.0.1_aarch64.apk
  ```

  ### x86_64

  ```shell
  curl -sL -o ./waterfall_1.0.1_x86_64.apk https://github.com/null93/waterfall/releases/download/1.0.1/waterfall_1.0.1_x86_64.apk
  apk add --allow-untrusted ./waterfall_1.0.1_x86_64.apk
  rm ./waterfall_1.0.1_x86_64.apk
  ```
</details>

<details>
  <summary>Arch</summary>
  
  ### aarch64

  ```shell
  curl -sL -o ./waterfall-1.0.1-1-aarch64.pkg.tar.zst https://github.com/null93/waterfall/releases/download/1.0.1/waterfall-1.0.1-1-aarch64.pkg.tar.zst
  sudo pacman -U ./waterfall-1.0.1-1-aarch64.pkg.tar.zst
  rm ./waterfall-1.0.1-1-aarch64.pkg.tar.zst
  ```

  ### x86_64

  ```shell
  curl -sL -o ./waterfall-1.0.1-1-x86_64.pkg.tar.zst https://github.com/null93/waterfall/releases/download/1.0.1/waterfall-1.0.1-1-x86_64.pkg.tar.zst
  sudo pacman -U ./waterfall-1.0.1-1-x86_64.pkg.tar.zst
  rm ./waterfall-1.0.1-1-x86_64.pkg.tar.zst
  ```
</details>

## TODO

- Improve details page
- Improve refresh function to cache already downloaded events
