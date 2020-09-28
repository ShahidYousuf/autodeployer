# autodeployer
### An automatic, one-install and configurable code deployment and code installation application. Utilizes features like local and remote repo tracking and automatic code installation and email notifications configurable with JSON syntax.

### Features
#### Local and Remote git repo tracking for changes.
#### Auto pull on remote repo updates
#### Auto install code based on config.json file listing commands to run in a flexible and standard format
#### Installation/Deployment Email notifications also configured via config.json file

### Setup
#### Install the binary application on Linux/Mac/Windows machine.
#### Set the following Environment variables
`export REPO_DIRECTORY_PATH="/path/to/your/repo/"`  
 `export REPO_NAME="repo"`  
 `export REPO_USER="bob"`  
 `export INSTALL_DIRECTIVES_PATH="/path/to/your/repo/config.json"`  
 `export GITHUB_ACCESS_TOKEN="some-access-token-here"`

#### Note that config.json file needs to inside the repo which controls the installation process and whether to actually run the installation or not.

#### config.json format
`
  "app_config": {`  
    `"run_installation": true,`  
    `"continue_on_fail": false,`  
    `"notify_emails": ["shahidyousuf77@gmail.com"]`  
  `},`  
  `"app_commands": [`  
    `{`  
      `"name": "List the items",`  
      `"run_dir": "/home/shahid/",`  
      `"command": ["ls"]`  
    `},`  
    `{`  
      `"name": "Activate the python environment",`  
      `"run_dir": "/home/shahid/",`  
      `"command": ["source", "/path/to/env/bin/activate"]`  
    `}`  
  `]`  
`}`  