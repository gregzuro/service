#!/bin/sh
cd /srv
DEPLOY_ENV=local ./switch master -l 7467 --short-name master1 --iga '[[-113,45],[-113,46],[-112,46],[-112,45]]' 