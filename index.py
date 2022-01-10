#!/usr/bin/env python

import os
import pathlib

from jinja2 import Environment, FileSystemLoader

def main():
    detail_pages = sorted([str(pathlib.Path(*p.parts[1:])) for p in
            list(pathlib.Path('html').glob('**/*.html'))])
    print(detail_pages)

    templates_dir = './templates/'
    env = Environment(loader = FileSystemLoader(templates_dir))
    template = env.get_template('index.html')

    filename = os.path.join('html', 'index.html')
    with open(filename, 'w') as fh:
        fh.write(template.render(links = detail_pages))

if __name__ == "__main__":
    main()
