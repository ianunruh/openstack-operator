#!/usr/bin/env python3
import os

import openstack


def main():
    username = os.environ['SVC_OS_USERNAME']
    password = os.environ['SVC_OS_PASSWORD']
    user_domain_name = os.environ['SVC_OS_USER_DOMAIN_NAME']

    project_name = os.environ.get('SVC_OS_PROJECT_NAME')
    project_domain_name = None
    if project_name:
        project_domain_name = os.environ['SVC_OS_PROJECT_DOMAIN_NAME']

    roles = os.environ.get('SVC_ROLES')
    if not roles:
        roles = 'admin'

    openstack.enable_logging(debug=True)
    conn = openstack.connect()

    user_kwargs = dict(password=password)

    # Ensure project if specified
    project = None
    if project_name:
        # Ensure project domain
        project_domain = conn.get_domain(name_or_id=project_domain_name)
        if not project_domain:
            project_domain = conn.create_domain(name=project_domain_name)

        project = conn.get_project(project_name, domain_id=project_domain.id)
        if not project:
            project = conn.create_project(name=project_name, domain_id=project_domain.id)

        user_kwargs['default_project'] = project.id

    # Ensure user domain
    domain = conn.get_domain(name_or_id=user_domain_name)
    if not domain:
        domain = conn.create_domain(name=user_domain_name)

    # Create or update user
    user = conn.get_user(username, domain_id=domain.id)
    if user:
        conn.update_user(user.id, domain_id=domain.id, **user_kwargs)
    else:
        user = conn.create_user(name=username, domain_id=domain.id, **user_kwargs)

    # Apply any roles to either project or domain
    for role_name in roles.split(','):
        role = conn.get_role(role_name)
        if not role:
            role = conn.create_role(name=role_name)

        if project:
            conn.grant_role(role, user=user.id, domain=domain.id, project=project.id)
        else:
            conn.grant_role(role, user=user.id, domain=domain.id)


if __name__ == '__main__':
    main()
