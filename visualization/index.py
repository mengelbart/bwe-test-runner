#!/usr/bin/env python

import os
import datetime
import json
import shutil

from glob import glob
from jinja2 import Environment, FileSystemLoader

def index(root, runs):
    keep = sorted(runs[-3:])
    remove = list(set(runs) - set(keep))
    print(keep)
    for r in remove:
        print('remove dir: ' + r)
        #shutil.rmtree(r)

#    dates = [os.path.basename(os.path.normpath(path)) for path in roots]
#    runs = sorted([datetime.datetime.fromtimestamp(int(s)).strftime('%Y-%m-%d %H:%M:%S.%f') for s in dates])

    templates_dir = './visualization/templates/'
    env = Environment( loader = FileSystemLoader(templates_dir) )
    template = env.get_template('index.html')

    filename = os.path.join(root, 'index.html')
    with open(filename, 'w') as fh:
        fh.write(template.render(
            href = os.path.basename(os.path.normpath(keep[-1])),
            ))
    




def run(root, runs):
    runs = [{
        'link': os.path.join('/', os.path.basename(os.path.normpath(root))),
        'name': datetime.datetime.fromtimestamp(int(root)).strftime('%Y-%m-%d %H:%M:%S'),
        } for root in runs]

    with open('cmd/testcases.json') as scenario_file:
        scenario_config = json.load(scenario_file)

    with open('cmd/implementations.json') as implementation_file:
        implementation_config = json.load(implementation_file)

    implementation_names = {os.path.basename(os.path.normpath(path)) for path in glob(root + '/*/')}
    scenario_names = {os.path.basename(os.path.normpath(path)) for path in glob(root + '/*/*/')}

    scenarios = sorted([{
        'id': int(scenario),
        'href': scenario_config[scenario]['href'],
        'description': scenario_config[scenario]['description']
        } for scenario in scenario_names], key=lambda d: d['id'])

    implementations = [{
        'name': implementation_config[implementation]['name'],
        'sender': {
            'href': implementation_config[implementation]['sender']['href'],
            'image': implementation_config[implementation]['sender']['image'],
            },
        'receiver': {
            'href': implementation_config[implementation]['receiver']['href'],
            'image': implementation_config[implementation]['receiver']['image'],
            },
        'scenarios': {
            int(os.path.basename(os.path.normpath(path))): os.path.join(implementation, os.path.basename(os.path.normpath(path)), 'detail.html')
            for path in glob(root + implementation + '/*/')},
        } for implementation in implementation_names]

    templates_dir = './visualization/templates/'
    env = Environment( loader = FileSystemLoader(templates_dir) )
    template = env.get_template('run.html')

    filename = os.path.join(root, 'index.html')
    with open(filename, 'w') as fh:
        fh.write(template.render(
            runs = runs,
            scenarios = scenarios,
            implementations = implementations,
            ))
    

def main():
    roots = glob('./gh-pages/*/')
    index('./gh-pages/', roots)

    for root in roots:
        run(root, [os.path.basename(os.path.normpath(run)) for run in roots])


if __name__ == "__main__":
    main()
