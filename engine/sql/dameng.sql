DROP TABLE IF EXISTS "RULEAIUSER.USER";
CREATE TABLE RULEAIUSER."USER"
(
    "id"          BIGINT IDENTITY(1, 1) NOT NULL,
    "username"    VARCHAR(256),
    "password"    VARCHAR(256),
    "create_time" TIMESTAMP,
    "update_time" TIMESTAMP,
    "enabled"     TINYINT DEFAULT 0,
    "deleted"     TINYINT DEFAULT 0,
    "role_id"     BIGINT,
    "org_id"      BIGINT,
    PRIMARY KEY ("id")
);

INSERT INTO RULEAIUSER."USER" ("username", PASSWORD, CREATE_TIME, UPDATE_TIME, ENABLED, DELETED, ROLE_ID, ORG_ID)
VALUES ('rule', '$2a$10$kS7stNgQ3zybAYtT5K2rLu9Utgf3EmUBsRDlw0fuK1CQ16fJhIRBu', '2023-10-20 05:04:06',
        '2023-10-20 05:04:06', 0, 0, 1, 0);


DROP TABLE IF EXISTS "RULEAIUSER.ROLE";

CREATE TABLE RULEAIUSER."ROLE"
(
    "id"          BIGINT IDENTITY(1, 1) NOT NULL,
    "rolename"    VARCHAR(256),
    "create_time" TIMESTAMP,
    "update_time" TIMESTAMP,
    "enabled"     TINYINT DEFAULT 0,
    "manager"     TINYINT DEFAULT 0,
    "deleted"     TINYINT DEFAULT 0,
    "org_id"      BIGINT,
    PRIMARY KEY (id)
);

INSERT INTO RULEAIUSER."ROLE" (ROLENAME, CREATE_TIME, UPDATE_TIME, ENABLED, MANAGER, DELETED, ORG_ID)
VALUES ('ROLE_ADMIN', '2023-10-20 05:04:06', '2023-10-20 05:04:06', 0, 1, 0, 0);


DROP TABLE IF EXISTS "RULEAIUSER.ROLE_GROUP";

CREATE TABLE RULEAIUSER.ROLE_GROUP
(
    "id"          BIGINT IDENTITY(1, 1) NOT NULL,
    "role_id"     BIGINT,
    "group_id"    BIGINT,
    "action"      VARCHAR(255),
    "create_time" TIMESTAMP,
    "update_time" TIMESTAMP,
    "deleted"     TINYINT DEFAULT 0,
    "org_id"      BIGINT,
    PRIMARY KEY (ID)
);


INSERT INTO RULEAIUSER.ROLE_GROUP (ROLE_ID, GROUP_ID, "ACTION", CREATE_TIME, UPDATE_TIME, DELETED, ORG_ID)
VALUES (1, 0, 'RW', '2023-10-20 05:04:06', '2023-10-20 05:04:06', 0, 0);


DROP TABLE IF EXISTS "RULEAIUSER.RULEAIUSER.RULE_DESIGN";

CREATE TABLE RULEAIUSER.RULE_DESIGN
(
    "id"           BIGINT IDENTITY(1, 1) NOT NULL,
    "code"         VARCHAR(255),
    "name"         VARCHAR(255),
    "desc"         VARCHAR(512),
    "design"       TEXT,
    "design_json"  CLOB,
    "version"      VARCHAR(255) DEFAULT '0',
    "uri"          VARCHAR(255),
    "storage_type" TINYINT      DEFAULT 0,
    "deleted"      TINYINT      DEFAULT 0,
    "publish"      TINYINT      DEFAULT 0,
    "create_time"  TIMESTAMP,
    "update_time"  TIMESTAMP,
    "group_id"     BIGINT       DEFAULT 0,
    "templated"    TINYINT      DEFAULT 0,
    "template_id"  BIGINT       DEFAULT 0,
    "type"         TINYINT      DEFAULT 0,
    PRIMARY KEY (ID)
);


DROP TABLE IF EXISTS "RULEAIUSER.RULE_DESIGN_VERSION";

CREATE TABLE RULEAIUSER.RULE_DESIGN_VERSION
(
    "id"             BIGINT IDENTITY(1, 1) NOT NULL,
    "rule_design_id" BIGINT,
    "design"         TEXT,
    "design_json"    CLOB,
    "version"        VARCHAR(255) DEFAULT '0',
    "uri"            VARCHAR(255),
    "publish"        TINYINT      DEFAULT 0,
    "create_time"    TIMESTAMP,
    "update_time"    TIMESTAMP,
    PRIMARY KEY (ID)
);


DROP TABLE IF EXISTS "RULEAIUSER.RULE_GROUP";

CREATE TABLE RULEAIUSER.RULE_GROUP
(
    "id"          BIGINT IDENTITY(1, 1) NOT NULL,
    "code"        VARCHAR(255),
    "name"        VARCHAR(255),
    "parent_id"   BIGINT  DEFAULT 0,
    "create_time" TIMESTAMP,
    "update_time" TIMESTAMP,
    "deleted"     TINYINT DEFAULT 0,
    "type"        TINYINT DEFAULT 0,
    PRIMARY KEY (ID)
);



DROP TABLE IF EXISTS "RULE_INFO";
CREATE TABLE RULEAIUSER.RULE_INFO
(
    "id"          BIGINT IDENTITY(1, 1) NOT NULL,
    "code"        VARCHAR(256),
    "name"        VARCHAR(255),
    "desc"        VARCHAR(512),
    "script"      TEXT,
    "create_time" TIMESTAMP,
    "update_time" TIMESTAMP,
    PRIMARY KEY (ID)
);