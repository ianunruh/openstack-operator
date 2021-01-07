#!/usr/bin/env python3
import os

import openstack


def main():
    username = os.environ['SVC_OS_USERNAME']
    password = os.environ['SVC_OS_PASSWORD']
    user_domain_name = os.environ['SVC_OS_USER_DOMAIN_NAME']
    project_name = os.environ['SVC_OS_PROJECT_NAME']
    project_domain_name = os.environ['SVC_OS_PROJECT_DOMAIN_NAME']

    openstack.enable_logging(debug=True)
    conn = openstack.connect()

    project_domain = conn.identity.find_domain(project_domain_name)
    user_domain = conn.identity.find_domain(user_domain_name)

    project = conn.identity.find_project(project_name, domain_id=project_domain.id)
    if not project:
        project = conn.identity.create_project(name=project_name, domain_id=project_domain.id)

    role = conn.identity.find_role('admin')

    user = conn.identity.find_user(username, domain_id=user_domain.id)
    if user:
        conn.identity.update_user(user, password=password, default_project_id=project.id)
    else:
        conn.identity.create_user(name=username, domain_id=user_domain.id, password=password, default_project_id=project.id)

    conn.identity.assign_project_role_to_user(project, user, role)


if __name__ == '__main__':
    main()
