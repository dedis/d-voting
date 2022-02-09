# Installation

Once the project cloned type `npm install` to install the packages.

[Install mysql on macOS](https://flaviocopes.com/mysql-how-to-install/)

# Database

This project needs a mysql database to work. This databse is needed for
the administration part (the roles)

## Connect to the mysql
```bash
 mysql -uroot
```

## Create the DB:
```
CREATE DATABASE dvotingtestdb;
USE dvotingtestdb;
SHOW DATABASES;
```

## Create the user:
```
CREATE USER  backend@localhost IDENTIFIED BY 'pa$$word';
GRANT ALL PRIVILEGES ON dvotingdb.* to backend@localhost;
```

## Create the table:
This database contains only one table that must be called `user_rights` and contains the following fields

- id : the primary key (it is an INT), it must have the auto-increment property
- sciper : INT
- role : VARCHAR(255), it is 255 by convention

```mysql
CREATE TABLE IF NOT EXISTS user_rights (
  id INT AUTO_INCREMENT PRIMARY KEY,
  sciper INT NOT NULL,
  role VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

# Config

The project contains one file that is not in git (because it is in the .gitignore).
This file is called `config.json` and is located at the root of the express project (like this file).

This files contains all the secrets and also the running information. It should be formatted this way :

```json
{
  "FRONT_END_URL" : "<url of the current site, this is used for the tequila callback>",
  "DELA_NODE_URL" : "<url of the dela node>",
  "SESSION_SECRET" : "<session secret>",
  "PUBLIC_KEY" : "<public key>",
  "PRIVATE_KEY" : "<private key>",
  "DB_USER" : "backend@localhost",
  "DB_PASS" : "pa$$word",
  "DB_DB" : "dvotingtestdb"
}
```

# Run the program

Once all the previous steps done, the project can be run using `npm start`
