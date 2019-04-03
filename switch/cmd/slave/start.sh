#!/bin/sh
cd /srv
DEPLOY_ENV=local ./switch slave -a $(mdata-get master) -p 7467 -l 7467 --short-name slave