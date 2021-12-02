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

def plot_rates(subdir, sender, plot_path):
    lr_df = pd.read_csv(
        os.path.join(subdir, 'router.log'),
        index_col = 0,
        names = ['time', 'bandwidth'],
        header = None,
        usecols = [0, 1],
    )
    lr_df.index = pd.to_datetime(lr_df.index - lr_df.index[0], unit='ms')

    cc_df = pd.read_csv(
        os.path.join(subdir, 'send_log/cc.log'),
        index_col = 0,
        names = ['time', 'target bitrate'],
        header = None,
        usecols = [0, 1],
    )
    cc_df.index = pd.to_datetime(cc_df.index - cc_df.index[0], unit='ms')

    # Hack to extend bandwidth limit step function plot until $cc_df.index.max() (end of plot)
    lr_last = lr_df.iloc[[-1]]
    lr_df = lr_df.append(lr_last)
    as_list = lr_df.index.tolist()
    as_list[-1] = cc_df.index.max()
    lr_df.index = as_list

    rtp_out_df = pd.read_csv(
        os.path.join(subdir, 'send_log/rtp_out.log'),
        index_col = 0,
        names = ['time', 'RTP bit/s sent'],
        header = None,
        usecols = [0, 6],
    )
    rtp_out_df.index = pd.to_datetime(rtp_out_df.index - rtp_out_df.index[0], unit='ms')
    rtp_out_df['RTP bit/s sent'] = rtp_out_df['RTP bit/s sent'].apply(lambda x: x * 8)
    rtp_out_df = rtp_out_df.resample('1s').sum()

    rtp_in_df = pd.read_csv(
        os.path.join(subdir, 'receive_log/rtp_in.log'),
        index_col = 0,
        names = ['time', 'RTP bit/s received'],
        header = None,
        usecols = [0, 6],
    )
    rtp_in_df.index = pd.to_datetime(rtp_in_df.index - rtp_in_df.index[0], unit='ms')
    rtp_in_df['RTP bit/s received'] = rtp_in_df['RTP bit/s received'].apply(lambda x: x * 8)
    rtp_in_df = rtp_in_df.resample('1s').sum()


    fig, ax = plt.subplots(figsize=(8,2), dpi=400)

    l2, = ax.plot(rtp_out_df.index, rtp_out_df.values, label='RTP sent')
    l3, = ax.plot(rtp_in_df.index, rtp_in_df.values, label='RTP received')
    l1, = ax.plot(cc_df.index, cc_df.values, label='Target Bitrate')
    l0, = ax.step(lr_df.index, lr_df.values, where='post', label='Bandwidth')

    plt.xlabel('time')
    plt.ylabel('rate')
    ax.legend(handles=[l0, l1, l2, l3])
    ax.xaxis.set_major_formatter(mdates.DateFormatter("%M:%S"))
    ax.yaxis.set_major_formatter(EngFormatter(unit='bit/s'))

    plt.savefig(os.path.join(plot_path, sender + '-plot.png'))

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

def main():
    with open('output/config.json') as config_file:
        config = json.load(config_file)

    run_id = str(config['date'])
    implementation = config['implementation']['name']
    testcase = config['scenario']['name']

    path = os.path.join('html', run_id, implementation, testcase)
    Path(path).mkdir(parents=True, exist_ok=True)

    match testcase:
        case "1":
            plot_rates('output/a', 'a', path)
            generate_html(path)
            copytree('output', os.path.join(path, 'log'), ignore=filter_empty)
        case ("2"|"3"):
            plot_rates('output/a', 'a', path)
            plot_rates('output/b', 'b', path)
            generate_html(path)
            copytree('output', os.path.join(path, 'log'))
        case _:
            print('invalid scenario')


if __name__ == "__main__":
    main()


