<p align="center">
  <a href="https://ergomake.dev">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://github.com/ergomake/ergomake/assets/6868147/0803a471-7d20-4f37-b092-4e77f223d500">
    <source media="(prefers-color-scheme: light)" srcset="https://github.com/ergomake/ergomake/assets/6868147/0353476d-27e0-4c70-8303-db4ee93aebef">
    <img alt="Ergomake logo" src="https://github.com/ergomake/ergomake/assets/6868147/0353476d-27e0-4c70-8303-db4ee93aebef">
    </picture>
  </a>
</p>

<h4 align="center">
  <a href="https://docs.ergomake.dev">Documentation</a> |
  <a href="https://ergomake.dev">Website</a>
</h4>

<p align="center">
  Preview environments on every pull-request, for any stack.
</p>
<p align="center">
  <a href="https://github.com/ergomake/ergomake/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/ergomake/ergomake" alt="Ergomake is released under the GNU GPLv3 license." />
  </a>
  <a href="https://discord.gg/daGzchUGDt">
    <img src="https://img.shields.io/badge/chat-on%20discord-7289DA.svg" alt="Discord Chat" />
  </a>
  <a href="https://twitter.com/intent/follow?screen_name=GetErgomake">
    <img src="https://img.shields.io/twitter/follow/GetErgomake.svg?label=Follow%20@GetErgomake" alt="Follow @GetErgomake" />
  </a>
</p>


## What is Ergomake

Every time you open a PR, Ergomake spins-up your entire application and sends you a preview link.

Anything that you can run in a container, you can run in Ergomake. Our previews may include your back-end, databases, and queues, for example.

Ergomake also supports multi-repo projects.

<p align="center">
  <img width="800" alt="intro" src="https://github.com/ergomake/ergomake/assets/6868147/b67f984e-f9c2-43bb-a780-b75671923aff">
</p>

## Getting Started

> You can see the complete documentation [here](https://docs.ergomake.dev/docs/intro).

1. [Log into Ergomake](https://app.ergomake.dev).
2. Select the desired organization and click the "Add Repo" button.
3. During the installation process, you'll be prompted to give it access to the repository for which you want to generate previews.
    **Make sure to select all the repositories you need**.

    > ⚠️ Ergomake can't generate previews if it doesn't have access to a repository.
4. Create a `docker-compose.yml` file in your repository's `.ergomake` folder, which should be in the repository's root.

    Ergomake will use this file to generate preview environments.

    ```yml
    # Here's an example docker-compose.yml file
    version: "3.8"
    services:
      # On pull requests, Ergomake can build your own images
      web:
        build: ..
        ports:
          - "8080:8080"

      # You can build a second repository by referencing a folder with
      # the desired repository name in a path *outside* your current repository.
      api:
        build: ../../my-backend-repo
        ports:
          - "3001:3001"

      database:
        image: mongo
        environment:
          MONGODB_INITDB_ROOT_USERNAME: username
          MONGODB_INITDB_ROOT_PASSWORD: password
    ```
5. Open a pull-request and wait for the Ergomake Bot's comment.
    That comment contains a link to all the applications running within your preview environment.


## Issues & Support

You can find Ergomake's users and maintainers in [GitHub Discussions](https://github.com/ergomake/ergomake/discussions). There you can ask how to set up Ergomake, ask us about the roadmap, and discuss any other related topics.

You can also reach us directly (and more quickly) in our [Discord server](https://discord.gg/daGzchUGDt).


## Other channels

- [Issue Tracker](https://github.com/ergomake/ergomake/issues)
- [Twitter](https://twitter.com/GetErgomake)
- [LinkedIn](https://www.linkedin.com/company/ergomake)
- [Ergomake Blog](https://ergomake.dev/blog)


## License

Licensed under the [GNU GPLv3 License](https://github.com/ergomake/ergomake/blob/main/LICENSE).
