#!/usr/bin/env python
import os
import json
import argparse
import pandas as pd
import matplotlib.pyplot as plt
import matplotlib.dates as mdates
import datetime as dt

from glob import glob
from matplotlib.ticker import PercentFormatter, EngFormatter
from jinja2 import Environment, FileSystemLoader

class rates_plot:
    def __init__(self, name):
        self.name = name
        self.labels = []
        self.fig, self.ax = plt.subplots(figsize=(8,3), dpi=400)

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
        l, = self.ax.plot(df.index, df.values, label=label, linewidth=0.5)
        self.labels.append(l)
        return True

    def add_rctp(self, file, basetime, label):
        if not os.path.exists(file):
            return False
        df = pd.read_csv(
                file,
                index_col = 0,
                names = ['time', 'rate'],
                header = None,
                usecols = [0, 1],
            )
        df.index = pd.to_datetime(df.index - basetime, unit='ms')
        df['rate'] = df['rate'].apply(lambda x: x * 8)
        df = df.resample('1s').sum()
        l, = self.ax.plot(df.index, df.values, label=label, linewidth=0.5)
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
        l, = self.ax.plot(df.index, df.values, label='Target Bitrate', linewidth=0.5)
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
        l, = self.ax.step(df.index, df.values, where='post', label='Capacity', linewidth=0.5)
        self.labels.append(l)
        return True

    def plot(self, path, prefix):
        self.ax.set_xlabel('Time')
        self.ax.set_ylabel('Rate')
        self.ax.set_title(self.name)
        self.ax.xaxis.set_major_formatter(mdates.DateFormatter("%M:%S"))
        self.ax.yaxis.set_major_formatter(EngFormatter(unit='bit/s'))
        self.ax.set_xlim([dt.datetime(1970, 1, 1), dt.datetime(1970, 1, 1,
            minute=1, second=45)])
        self.ax.set_ylim([0, 2600000])
        lgd = self.ax.legend(handles = self.labels, loc='upper right', bbox_to_anchor=(0.75, 1))
        #lgd = self.ax.legend(handles = self.labels, loc='upper center', bbox_to_anchor=(0.5, -0.5), ncol=len(self.labels))
        self.fig.tight_layout()

        self.fig.savefig(os.path.join(path, prefix + '.png'), bbox_extra_artists=(lgd,), bbox_inches='tight')

class gcc_plot:
    def __init__(self, name):
        self.name = name
        self.rtt_fig, self.rtt_ax = plt.subplots(figsize=(8,2), dpi=400)
        self.rtt_labels = []
        self.estimate_fig, self.estimate_ax = plt.subplots(figsize=(8,2), dpi=400)
        self.estimate_labels = []
        self.loss_fig, self.loss_ax = plt.subplots(figsize=(8,2), dpi=400)
        self.loss_labels = []

    def add(self, file, basetime):
        if not os.path.exists(file):
            return False
        df = pd.read_csv(
                file,
                index_col = 0,
                names = ['time', 'averageLoss', 'measurement', 'estimate', 'threshold', 'rtt'],
                header = None,
                usecols = [0, 3, 5, 6, 7, 8]
            )
        df.index = pd.to_datetime(df.index - basetime, unit='ms')

        l, = self.rtt_ax.plot(df.index, df['rtt'], label='RTT', linewidth=0.5)
        self.rtt_labels.append(l)

        l0, = self.estimate_ax.plot(df.index, df['estimate'], label='Estimate', linewidth=0.5)
        l1, = self.estimate_ax.plot(df.index, df['threshold'], label='Threshold', linewidth=0.5)
        l2, = self.estimate_ax.plot(df.index, -df['threshold'], label='-Threshold', linewidth=0.5)
        l3, = self.estimate_ax.plot(df.index, df['measurement'], label='Measurement', linewidth=0.5)
        self.estimate_labels.append(l0)
        self.estimate_labels.append(l1)
        self.estimate_labels.append(l2)
        self.estimate_labels.append(l3)

        l, = self.loss_ax.plot(df.index, df['averageLoss'], label='Average Loss', linewidth=0.5)
        self.loss_labels.append(l)
        return True

    def plot(self, path, prefix):
        self.rtt_ax.set_xlabel('Time')
        self.rtt_ax.set_ylabel('RTT')
        self.rtt_ax.set_title(self.name + ' RTT')
        self.rtt_ax.xaxis.set_major_formatter(mdates.DateFormatter("%M:%S"))
        self.rtt_ax.yaxis.set_major_formatter(EngFormatter(unit='ms'))
        self.rtt_ax.set_xlim([dt.datetime(1970, 1, 1), dt.datetime(1970, 1, 1, minute=1, second=45)])
        lgd = self.rtt_ax.legend(handles = self.rtt_labels, loc='upper right', bbox_to_anchor=(0.75, 1))
        self.rtt_fig.savefig(os.path.join(path, prefix + '-rtt.png'), bbox_extra_artists=(lgd,), bbox_inches='tight')

        self.estimate_ax.set_xlabel('Time')
        self.estimate_ax.set_ylabel('Estimate')
        self.estimate_ax.set_title(self.name + ' Estimate')
        self.estimate_ax.xaxis.set_major_formatter(mdates.DateFormatter("%M:%S"))
        self.estimate_ax.yaxis.set_major_formatter(EngFormatter(unit='ms'))
        self.estimate_ax.set_xlim([dt.datetime(1970, 1, 1), dt.datetime(1970, 1, 1, minute=1, second=45)])
        lgd = self.estimate_ax.legend(handles = self.estimate_labels, loc='upper right', bbox_to_anchor=(0.75, 1))
        self.estimate_fig.tight_layout()
        self.estimate_fig.savefig(os.path.join(path, prefix + '-estimate.png'), bbox_extra_artists=(lgd,), bbox_inches='tight')

        self.loss_ax.set_xlabel('Time')
        self.loss_ax.set_ylabel('Loss')
        self.loss_ax.set_title(self.name + ' Loss')
        self.loss_ax.xaxis.set_major_formatter(mdates.DateFormatter("%M:%S"))
        self.loss_ax.yaxis.set_major_formatter(PercentFormatter(xmax=1.0))
        self.loss_ax.set_xlim([dt.datetime(1970, 1, 1), dt.datetime(1970, 1, 1, minute=1, second=45)])
        lgd = self.loss_ax.legend(handles = self.loss_labels, loc='upper right', bbox_to_anchor=(0.75, 1))
        self.loss_fig.tight_layout()
        self.loss_fig.savefig(os.path.join(path, prefix + '-loss.png'), bbox_extra_artists=(lgd,), bbox_inches='tight')

class scream_plot:
    def __init__(self, name):
        self.name = name
        self.delay_fig, self.delay_ax = plt.subplots(figsize=(8,2), dpi=400)
        self.rates_fig, self.rates_ax = plt.subplots(figsize=(8,2), dpi=400)
        self.in_flight_fig, self.in_flight_ax = plt.subplots(figsize=(8,2), dpi=400)
        self.delay_labels = []
        self.rates_labels = []
        self.in_flight_labels = []

    def add_scream(self, file, basetime):
        if not os.path.exists(file):
            return False

        df = pd.read_csv(
                file,
                index_col = 0,
                names = ['time', 'queue_delay', 'srtt', 'cwnd', 'bytes_in_flight',
                    'rate_lost', 'rate_transmitted', 'rate_acked'],
                header = None,
                usecols = [0, 2, 3, 4, 5, 6, 7, 8],
            )
        df.index = pd.to_datetime(df.index - basetime, unit='ms')
        df['rate_lost'] = pd.to_numeric(df['rate_lost'], errors='coerce')
        df['rate_transmitted'] = pd.to_numeric(df['rate_transmitted'], errors='coerce')
        df['rate_acked'] = pd.to_numeric(df['rate_acked'], errors='coerce')

        l0, = self.delay_ax.plot(df.index, df['queue_delay'], label='Queue Delay', linewidth=0.5)
        #l1, = self.delay_ax.plot(df.index, df['srtt'], label='SRTT', linewidth=0.5)
        self.delay_labels.append(l0)
        #self.delay_labels.append(l1)

        l0, = self.in_flight_ax.plot(df.index, df['cwnd'], label='CWND', linewidth=0.5)
        l1, = self.in_flight_ax.plot(df.index, df['bytes_in_flight'], label='Bytes in Flight', linewidth=0.5)
        self.in_flight_labels.append(l0)
        self.in_flight_labels.append(l1)

        l0, = self.rates_ax.plot(df.index, df['rate_lost'], label='Rate lost', linewidth=0.5)
        l1, = self.rates_ax.plot(df.index, df['rate_transmitted'], label='Rate transmitted', linewidth=0.5)
        l2, = self.rates_ax.plot(df.index, df['rate_acked'], label='Rate acked', linewidth=0.5)
        self.rates_labels.append(l0)
        self.rates_labels.append(l1)
        self.rates_labels.append(l2)
        return True

    def plot(self, path, prefix):
        self.in_flight_ax.set_xlabel('Time')
        self.in_flight_ax.set_ylabel('In flight')
        self.in_flight_ax.set_title('Bytes in Flight ' + self.name)
        self.in_flight_ax.xaxis.set_major_formatter(mdates.DateFormatter("%M:%S"))
        self.in_flight_ax.yaxis.set_major_formatter(EngFormatter(unit='Byte'))
        self.in_flight_ax.set_xlim([dt.datetime(1970, 1, 1), dt.datetime(1970, 1, 1, minute=1, second=45)])
        lgd = self.in_flight_ax.legend(handles = self.in_flight_labels, loc='upper right', bbox_to_anchor=(0.75, 1))
        self.in_flight_fig.tight_layout()
        self.in_flight_fig.savefig(os.path.join(path, prefix + '-in-flight.png'), bbox_extra_artists=(lgd,), bbox_inches='tight')

        self.rates_ax.set_xlabel('Time')
        self.rates_ax.set_ylabel('Rate')
        self.rates_ax.set_title('Rates ' + self.name)
        self.rates_ax.xaxis.set_major_formatter(mdates.DateFormatter("%M:%S"))
        self.rates_ax.yaxis.set_major_formatter(EngFormatter(unit='kbit/s'))
        self.rates_ax.set_xlim([dt.datetime(1970, 1, 1), dt.datetime(1970, 1, 1, minute=1, second=45)])
        lgd = self.rates_ax.legend(handles = self.rates_labels, loc='upper right', bbox_to_anchor=(0.75, 1))
        self.rates_fig.tight_layout()
        self.rates_fig.savefig(os.path.join(path, prefix + '-rates.png'), bbox_extra_artists=(lgd,), bbox_inches='tight')

        self.delay_ax.set_xlabel('Time')
        self.delay_ax.set_ylabel('Delay')
        self.delay_ax.set_title('Delay ' + self.name)
        self.delay_ax.xaxis.set_major_formatter(mdates.DateFormatter("%M:%S"))
        self.delay_ax.yaxis.set_major_formatter(EngFormatter(unit='s'))
        self.delay_ax.set_xlim([dt.datetime(1970, 1, 1), dt.datetime(1970, 1, 1, minute=1, second=45)])
        lgd = self.delay_ax.legend(handles = self.delay_labels, loc='upper right', bbox_to_anchor=(0.75, 1))
        self.delay_fig.tight_layout()
        self.delay_fig.savefig(os.path.join(path, prefix + '-delay.png'), bbox_extra_artists=(lgd,), bbox_inches='tight')

class qlog_cwnd_plot:
    def __init__(self, name):
        self.name = name
        self.fig, self.ax = plt.subplots(figsize=(8,2), dpi=400)
        self.labels = []

    def add_cwnd(self, file):
        if not os.path.exists(file):
            return False

        inflight = []
        congestion = []
        with open(file) as f:
            for index, line in enumerate(f):
                event = json.loads(line.strip())
                if 'name' in event and event['name'] == 'recovery:metrics_updated':
                    if 'data' in event and 'bytes_in_flight' in event['data']:
                        inflight.append({'time': event['time'], 'bytes_in_flight': event['data']['bytes_in_flight']})
                if 'name' in event and event['name'] == 'recovery:metrics_updated':
                    if 'data' in event and 'congestion_window' in event['data']:
                        congestion.append({'time': event['time'], 'cwnd': event['data']['congestion_window']})

        df = pd.DataFrame(inflight)
        df.index = pd.to_datetime(df['time'], unit='ms')
        l0, = self.ax.plot(df.index, df['bytes_in_flight'], label='Bytes in Flight', linewidth=0.5)
        self.labels.append(l0)

        df = pd.DataFrame(congestion)
        df.index = pd.to_datetime(df['time'], unit='ms')
        l, = self.ax.plot(df.index, df['cwnd'], label='CWND', linewidth=0.5)
        self.labels.append(l)
        return True

    def plot(self, path, prefix):
        self.ax.set_xlabel('Time')
        self.ax.set_ylabel('CWND')
        self.ax.set_title(self.name)
        self.ax.yaxis.set_major_formatter(EngFormatter(unit='Bytes'))
        self.ax.xaxis.set_major_formatter(mdates.DateFormatter("%M:%S"))
        self.ax.set_xlim([dt.datetime(1970, 1, 1), dt.datetime(1970, 1, 1, minute=1, second=45)])
        lgd = self.ax.legend(handles = self.labels, loc='upper right', bbox_to_anchor=(0.75, 1))
        self.fig.tight_layout()
        self.fig.savefig(os.path.join(path, prefix + '.png'), bbox_extra_artists=(lgd,), bbox_inches='tight')

class qlog_bytes_sent_plot:
    def __init__(self, name):
        self.name = name
        self.fig, self.ax = plt.subplots(figsize=(9,4), dpi=400)
        self.labels = []

    def add_bytes_sent(self, file):
        if not os.path.exists(file):
            return False

        dgram = []
        stream = []
        sums = []
        with open(file) as f:
            for index, line in enumerate(f):
                event = json.loads(line.strip())
                if 'name' in event and event['name'] == 'transport:packet_sent':
                    if 'data' in event and 'frames' in event['data']:
                        datagrams = [frame for frame in event['data']['frames'] if frame['frame_type'] == 'datagram' ]
                        stream_frames = [frame for frame in event['data']['frames'] if frame['frame_type'] == 'stream' ]
                        if len(datagrams) > 0:
                            dgram.append({'time': event['time'], 'bytes': sum([datagram['length'] for datagram in datagrams])})
                            sums.append({'time': event['time'], 'bytes': sum([datagram['length'] for datagram in datagrams])})
                        if len(stream_frames) > 0:
                            stream.append({'time': event['time'], 'bytes': sum([stream['length'] for stream in stream_frames])})
                            sums.append({'time': event['time'], 'bytes': sum([stream['length'] for stream in stream_frames])})

        if len(dgram) > 0:
            datagram_df = pd.DataFrame(dgram)
            datagram_df.index = pd.to_datetime(datagram_df['time'], unit='ms')
            datagram_df['bytes'] = datagram_df['bytes'].apply(lambda x: x * 8)
            datagram_df = datagram_df.resample('1s').sum()
            l1, = self.ax.plot(datagram_df.index, datagram_df['bytes'], label='Datagram Sent', linewidth=0.5)
            self.labels.append(l1)

        if len(stream) > 0:
            stream_df = pd.DataFrame(stream)
            stream_df.index = pd.to_datetime(stream_df['time'], unit='ms')
            stream_df['bytes'] = stream_df['bytes'].apply(lambda x: x * 8)
            stream_df = stream_df.resample('1s').sum()
            l2, = self.ax.plot(stream_df.index, stream_df['bytes'], label='Stream Sent', linewidth=0.5)
            self.labels.append(l2)

        if len(sums) > 0:
            sums_df = pd.DataFrame(sums)
            sums_df.index = pd.to_datetime(sums_df['time'], unit='ms')
            sums_df['bytes'] = sums_df['bytes'].apply(lambda x: x * 8)
            sums_df = sums_df.resample('1s').sum()
            l3, = self.ax.plot(sums_df.index, sums_df['bytes'], label='Total sent', linewidth=0.5)
            self.labels.append(l3)

        return True

    def plot(self, path, prefix):
        self.ax.set_xlabel('Time')
        self.ax.set_ylabel('Bytes in Flight')

        self.ax.set_title(self.name)
        self.ax.xaxis.set_major_formatter(mdates.DateFormatter("%M:%S"))
        self.ax.yaxis.set_major_formatter(EngFormatter(unit='bit/s'))
        self.ax.set_xlim([dt.datetime(1970, 1, 1), dt.datetime(1970, 1, 1, minute=1, second=45)])

        lgd = self.ax.legend(handles = self.labels, loc='upper right', bbox_to_anchor=(0.75, 1))
        self.fig.tight_layout()
        self.fig.savefig(os.path.join(path, prefix + '.png'), bbox_extra_artists=(lgd,), bbox_inches='tight')

class qlog_rtt_plot:
    def __init__(self, name):
        self.name = name
        self.fig, self.ax = plt.subplots(figsize=(8,2), dpi=400)
        self.labels = []

    def add_rtt(self, file):
        if not os.path.exists(file):
            return False

        rtt = []
        with open(file) as f:
            for index, line in enumerate(f):
                event = json.loads(line.strip())
                if 'name' in event and event['name'] == 'recovery:metrics_updated':
                    append = False
                    sample = {'time': event['time']}
                    if 'data' in event and 'smoothed_rtt' in event['data']:
                        sample['smoothed_rtt'] = event['data']['smoothed_rtt']
                        append = True
                    if 'data' in event and 'min_rtt' in event['data']:
                        sample['min_rtt'] = event['data']['min_rtt']
                        append = True
                    if 'data' in event and 'latest_rtt' in event['data']:
                        sample['latest_rtt'] = event['data']['latest_rtt']
                        append = True
                    if append:
                        rtt.append(sample)

        df = pd.DataFrame(rtt)
        df.index = pd.to_datetime(df['time'], unit='ms')
        l, = self.ax.plot(df.index, df['latest_rtt'], label='Latest RTT', linewidth=0.5)
        self.labels.append(l)
        return True

    def plot(self, path, prefix):
        self.ax.set_xlabel('Time')
        self.ax.set_ylabel('RTT')

        self.ax.set_title(self.name)
        self.ax.xaxis.set_major_formatter(mdates.DateFormatter("%M:%S"))
        self.ax.yaxis.set_major_formatter(EngFormatter(unit='ms'))
        self.ax.set_xlim([dt.datetime(1970, 1, 1), dt.datetime(1970, 1, 1, minute=1, second=45)])

        lgd = self.ax.legend(handles = self.labels, loc='upper right', bbox_to_anchor=(0.75, 1))
        self.fig.tight_layout()
        self.fig.savefig(os.path.join(path, prefix + '.png'), bbox_extra_artists=(lgd,), bbox_inches='tight')

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
        l, = self.ax.step(df.index, df.values, where='post', label='Bandwidth', linewidth=0.5)
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

    def plot(self, path, prefix):
        self.ax.set_xlabel('Time')
        self.ax.set_ylabel('Rate')
        self.ax.set_title(self.name)
        self.ax.xaxis.set_major_formatter(mdates.DateFormatter("%M:%S"))
        self.ax.yaxis.set_major_formatter(EngFormatter(unit='bit/s'))
        self.ax.set_xlim([dt.datetime(1970, 1, 1), dt.datetime(1970, 1, 1, minute=1, second=45)])

        lgd = self.ax.legend(handles = self.labels, loc='upper right', bbox_to_anchor=(0.75, 1))
        self.fig.tight_layout()
        self.fig.savefig(os.path.join(path, prefix + '.png'), bbox_extra_artists=(lgd,), bbox_inches='tight')

def generate_html(path):
    images = [os.path.basename(x) for x in glob(os.path.join(path, '*.png'))]
    templates_dir = './templates/'
    env = Environment(loader = FileSystemLoader(templates_dir))
    template = env.get_template('detail.html')

    filename = os.path.join(path, 'index.html')
    with open(filename, 'w') as fh:
        fh.write(template.render(images = images))

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("plot")

    parser.add_argument("--name")
    parser.add_argument("--input_dir")
    parser.add_argument("--output_dir")
    parser.add_argument("--basetime", type=int, default=0)
    parser.add_argument("--router")

    args = parser.parse_args()

    match args.plot:
        case 'rates':
            basetime = pd.to_datetime(args.basetime, unit='s').timestamp() * 1000
            plot = rates_plot('RTP Rates ' + args.name)
            plot.add_rtp(os.path.join(args.input_dir, 'send_log', 'rtp_out.log'), basetime, 'RTP sent')
            plot.add_rtp(os.path.join(args.input_dir, 'receive_log', 'rtp_in.log'), basetime, 'RTP received')
            plot.add_cc(os.path.join(args.input_dir, 'send_log', 'gcc.log'), basetime)
            plot.add_cc(os.path.join(args.input_dir, 'send_log', 'scream.log'), basetime)
            plot.add_cc(os.path.join(args.input_dir, 'send_log', 'cc.log'), basetime)
            plot.add_router(args.router, basetime)
            plot.plot(args.output_dir, args.name + '-rtp-' + args.plot)

            plot = rates_plot('RTCP Rates ' + args.name)
            plot.add_rctp(os.path.join(args.input_dir, 'receive_log', 'rtcp_out.log'), basetime, 'RTCP sent')
            plot.add_rctp(os.path.join(args.input_dir, 'send_log', 'rtcp_in.log'), basetime, 'RTCP received')
            plot.plot(args.output_dir, args.name + '-rtcp-' + args.plot)

        case 'gcc':
            basetime = pd.to_datetime(args.basetime, unit='s').timestamp() * 1000
            plot = gcc_plot('GCC ' + args.name)
            if plot.add(os.path.join(args.input_dir, 'send_log', 'gcc.log'), basetime):
                plot.plot(args.output_dir, args.name + '-gcc')

        case 'scream':
            basetime = pd.to_datetime(args.basetime, unit='s').timestamp() * 1000
            plot = scream_plot('SCReAM ' + args.name)
            if plot.add_scream(os.path.join(args.input_dir, 'send_log', 'scream.log'), basetime):
                plot.plot(args.output_dir, args.name + '-scream')

        case 'qlog-cwnd':
            basetime = pd.to_datetime(args.basetime, unit='s').timestamp() * 1000
            qlog_files = glob(os.path.join(args.input_dir, 'send_log', '*.qlog'))
            plot = qlog_cwnd_plot('QLOG Sender CWND' + args.name)
            if len(qlog_files) > 0:
                if plot.add_cwnd(qlog_files[0]):
                    plot.plot(args.output_dir, args.name + '-qlog-sender-cwnd')

            qlog_files = glob(os.path.join(args.input_dir, 'receive_log', '*.qlog'))
            plot = qlog_cwnd_plot('QLOG Receiver CWND' + args.name)
            if len(qlog_files) > 0:
                if plot.add_cwnd(qlog_files[0]):
                    plot.plot(args.output_dir, args.name + '-qlog-receiver-cwnd')

        case 'qlog-bytes-sent':
            basetime = pd.to_datetime(args.basetime, unit='s').timestamp() * 1000
            qlog_files = glob(os.path.join(args.input_dir, 'send_log', '*.qlog'))
            plot = qlog_bytes_sent_plot('QLOG Sender Bytes Sent' + args.name)
            if len(qlog_files) > 0:
                if plot.add_bytes_sent(qlog_files[0]):
                    plot.plot(args.output_dir, args.name + '-qlog-sender-sent')

            qlog_files = glob(os.path.join(args.input_dir, 'receive_log', '*.qlog'))
            plot = qlog_bytes_sent_plot('QLOG Receiver Bytes Sent' + args.name)
            if len(qlog_files) > 0:
                if plot.add_bytes_sent(qlog_files[0]):
                    plot.plot(args.output_dir, args.name + '-qlog-receiver-sent')

        case 'qlog-rtt':
            basetime = pd.to_datetime(args.basetime, unit='s').timestamp() * 1000
            qlog_files = glob(os.path.join(args.input_dir, 'send_log', '*.qlog'))
            plot = qlog_rtt_plot('QLOG Sender RTT ' + args.name)
            if len(qlog_files) > 0:
                if plot.add_rtt(qlog_files[0]):
                    plot.plot(args.output_dir, args.name + '-qlog-sender-rtt')

            qlog_files = glob(os.path.join(args.input_dir, 'receive_log', '*.qlog'))
            plot = qlog_rtt_plot('QLOG Receiver RTT ' + args.name)
            if len(qlog_files) > 0:
                if plot.add_rtt(qlog_files[0]):
                    plot.plot(args.output_dir, args.name + '-qlog-receiver-rtt')

        case 'tcp':
            basetime = pd.to_datetime(args.basetime, unit='s').timestamp() * 1000
            plot = tcp_plot('TCP ' + args.name)
            plot.add_router(args.router, basetime)
            found_sent = plot.add(os.path.join(args.input_dir, 'send_log', 'tcp.log'), 'TCP sent')
            found_received = plot.add(os.path.join(args.input_dir, 'receive_log', 'tcp.log'), 'TCP received')
            if found_sent or found_received:
                plot.plot(args.output_dir, args.name + '-tcp')

        case 'html':
            generate_html(args.output_dir)

if __name__ == "__main__":
    main()


