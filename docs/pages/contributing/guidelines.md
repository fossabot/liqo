---
title: "Contributing Guidelines"
---

:+1::tada: First off, thank you for taking the time to contribute to Liqo! :tada::+1:

The following is a set of guidelines for contributing to Liqo, which are hosted in the LiqoTech Organization on GitHub. 
These are mostly guidelines, not rules. Use your best judgment, and feel free to propose changes to this document
in a pull request.

## Your First Code Contribution

Unsure where to begin contributing to Liqo? You can start by looking through the help-wanted issues.

# Repository Structure

The liqo repository structure follows the [Golang standard layout](https://github.com/golang-standards/project-layout). 
All components have the same structure:

If you want to read about using Liqo or developing packages in Liqo, the Liqo Flight Manual is free and available online. 

## Local development

Liqo components can be developed locally. We provide a [deployment script](/scripts/kind/kind.sh) to spawn multiple 
kubernetes clusters by using [Kind](https://kind.sigs.k8s.io/) with Liqo installed. This script can be used as a starting
point to improve/replace one Liqo component.

## Pull Requests

The process described here has several goals:

* Maintain and possibly improve Liqo's quality
* Fix problems that are important to users
* Engage the community in working toward the best possible Liqo and to embrace new possible use-cases

## Styleguides

### Git Commit Messages

* Use the present tense ("Add feature" not "Added feature")
* Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
* Limit the first line to 72 characters or less
* Reference issues and pull requests liberally after the first line

#### Credits

[Liqo Contributing Guidelines](https://github.com/atom/atom/blob/master/CONTRIBUTING.md) inspired us when writing this 
document. Many thanks!