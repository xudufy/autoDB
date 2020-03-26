
use autodb;


INSERT INTO users(email, pw, username)
VALUES ('km19@gmail.com','1234567890','km19');

INSERT INTO projects(pname)
VALUES ('project1');

INSERT INTO project_developer(pid, pname, privilege)
VALUES (1, 1, 'owner'); /* assuming the auto_increment starts at 1 */

INSERT INTO tables(pid, name)
VALUES (1, 'table1'); /* assuming the auto_increment starts at 1 */

INSERT INTO apis(aid, tid, name, type, tmpl)
VALUES ('apiIDfromsha256',1,'table1','public','1234567890');