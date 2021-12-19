#!/usr/bin/env python
import os
import json
import pandas as pd
import matplotlib.pyplot as plt
import matplotlib.dates as mdates

from shutil import copytree
from pathlib import Path
from matplotlib.ticker import EngFormatter
from jinja2 import Environment, FileSystemLoader

def generate_html(path):
    templates_dir = './visualization/templates/'
    env = Environment( loader = FileSystemLoader(templates_dir) )
    template = env.get_template('detail.html')

    filename = os.path.join(path, 'detail.html')
    with open(filename, 'w') as fh:
        fh.write(template.render())

def filter_empty(dir, content):
    return [path for path in content if os.path.isdir(os.path.join(dir, path))
            and len(os.listdir(os.path.join(dir, path))) == 0]

class rates_plot:
    def __init__(self, name):
        self.name = name
        self.labels = []
        self.fig, self.ax = plt.subplots(figsize=(8,2), dpi=400)

    def add_rtp(self, file, basetime, label):
        if not os.path.exists(file):
            return False
        df = pd.read_csv(
                file,
                index_col = 0,
                names = ['time', 'rate'],
                header = None,
                usecols = [0, 6],
            )
        df.index = pd.to_datetime(df.index - basetime, unit='ms')
        df['rate'] = df['rate'].apply(lambda x: x * 8)
        df = df.resample('1s').sum()
        l, = self.ax.plot(df.index, df.values, label=label)
        self.labels.append(l)
        return True

    def add_cc(self, file, basetime):
        if not os.path.exists(file):
            return False
        df = pd.read_csv(
                file,
                index_col = 0,
                names = ['time', 'target'],
                header = None,
                usecols = [0, 1],
            )
        df.index = pd.to_datetime(df.index - basetime, unit='ms')
        df = df[df['target'] > 0]
        l, = self.ax.plot(df.index, df.values, label='Target Bitrate')
        self.labels.append(l)
        return True

    def add_router(self, file, basetime):
        if not os.path.exists(file):
            return False

        df = pd.read_csv(
                file,
                index_col = 0,
                names = ['time', 'bandwidth'],
                header = None,
                usecols = [0, 1],
            )
        df.index = pd.to_datetime(df.index - basetime, unit='ms')
        l, = self.ax.step(df.index, df.values, where='post', label='Bandwidth')
        self.labels.append(l)
        return True

    def plot(self, path):
        plt.xlabel('time')
        plt.ylabel('rate')
        self.ax.legend(handles=self.labels)
        self.ax.xaxis.set_major_formatter(mdates.DateFormatter("%M:%S"))
        self.ax.yaxis.set_major_formatter(EngFormatter(unit='bit/s'))

        plt.savefig(os.path.join(path, self.name + '-plot.png'))

class tcp_plot:
    def __init__(self, name):
        self.name = name
        self.fig, self.ax = plt.subplots(figsize=(8,2), dpi=400)
        self.labels = []

    def add_router(self, file, basetime):
        if not os.path.exists(file):
            return False

        df = pd.read_csv(
                file,
                index_col = 0,
                names = ['time', 'bandwidth'],
                header = None,
                usecols = [0, 1],
            )
        df.index = pd.to_datetime(df.index - basetime, unit='ms')
        l, = self.ax.step(df.index, df.values, where='post', label='Bandwidth')
        self.labels.append(l)
        return True

    def add(self, file, label):
        if not os.path.exists(file):
            return False

        with open(file) as data_file:
            data = json.load(data_file)

        df = pd.json_normalize(data, record_path='intervals')
        df.index = pd.to_datetime(df['sum.start'], unit='s')
        df = df.resample('1s').mean()

        l, = self.ax.plot(df.index, df['sum.bits_per_second'], label=label, linewidth=0.5)
        self.labels.append(l)
        return True

    def plot(self, path):
        plt.xlabel('Time')
        plt.ylabel('Rate')

        self.ax.legend(handles=self.labels)
        self.ax.xaxis.set_major_formatter(mdates.DateFormatter("%M:%S"))
        self.ax.yaxis.set_major_formatter(EngFormatter(unit='bit/s'))

        plt.savefig(os.path.join(path, self.name + '-plot.png'))



def main():
    sources = ['a', 'b', 'c']
    output_dir = 'output'
    html_dir = 'html'

    with open('output/config.json') as config_file:
        config = json.load(config_file)

    run_id = str(config['date'])
    implementation = config['implementation']['name']
    testcase = config['scenario']['name']

    path = os.path.join(html_dir, run_id, implementation, testcase)
    Path(path).mkdir(parents=True, exist_ok=True)

    copytree('output', os.path.join(path, 'log'), ignore=filter_empty)
    generate_html(path)

    basetime = pd.to_datetime(run_id, unit='s').timestamp() * 1000

    for source in sources:
        dir = os.path.join(output_dir, source)
        if os.path.isdir(dir):
            plot = rates_plot(source)

            found_log = False
            if plot.add_cc(os.path.join(dir, 'send_log', 'cc.log'), basetime):
                found_log = True

            if plot.add_rtp(os.path.join(dir, 'send_log',
                'rtp_out.log'), basetime, 'RTP received'):
                found_log = True

            if plot.add_rtp(os.path.join(dir, 'receive_log',
                'rtp_in.log'), basetime, 'RTP sent'):
                found_log = True

            if found_log:
                plot.add_router('output/leftrouter.log', basetime)
                plot.plot(os.path.join(path))

    tcp_receive_log = os.path.join(output_dir, 'tcp', 'receive_log', 'tcp.log')
    tcp_send_log = os.path.join(output_dir, 'tcp', 'send_log', 'tcp.log')
    tcp = tcp_plot('tcp')
    found_tcp = False
    if tcp.add(tcp_receive_log, 'TCP received'):
        found_tcp = True
    else:
        print("NO RECEIVELOG: " + tcp_receive_log)
    if tcp.add(tcp_send_log, 'TCP sent'):
        found_tcp = True
    else:
        print("NO SENDLOG: " + tcp_send_log)

    if found_tcp:
        tcp.add_router('output/leftrouter.log', basetime)
        tcp.plot(path)

if __name__ == "__main__":
    main()


