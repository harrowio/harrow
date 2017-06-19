production:
	ANSIBLE_CONFIG=./ansible.cfg ansible-playbook --inventory=./inventories/production production.yml

development:
	ANSIBLE_CONFIG=./ansible.cfg vagrant up dev --provision

appliance:
	ANSIBLE_CONFIG=./ansible.cfg vagrant up appliance --provision

.PHONY: production development enterprise
