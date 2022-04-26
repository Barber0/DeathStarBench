#!/usr/bin/env python3
import pathlib
import subprocess
import time

from subprocess import DEVNULL, PIPE

components = [
    'unique-id-service',
    'user-mention-service',
    'user-service',
    'user-timeline-service',
    'social-graph-service',
    'text-service',
    'url-shorten-service',
    'post-storage-service',
    'media-service',
    'compose-post-service',
    'home-timeline-service',
    'post-storage-mongodb',
    'media-memcached',
    'media-mongodb',
    'post-storage-memcached',
    'social-graph-redis',
    'user-timeline-mongodb',
    'compose-post-redis',
    'nginx-thrift',
    'write-home-timeline-rabbitmq',
    'home-timeline-redis',
    'url-shorten-memcached',
    'user-mongodb',
    'media-frontend',
    'user-timeline-redis',
    'user-memcached',
    'social-graph-mongodb',
    'url-shorten-mongodb',
    'write-home-timeline-service',
]


def stat_path(qos, uid, stat):
    family, _, name = stat.partition('.')
    slices = f'kubepods.slice/kubepods-{qos}.slice/kubepods-{qos}-pod{uid.replace("-", "_")}.slice'
    return pathlib.Path(f'/sys/fs/cgroup/{family}/{slices}/{family}.{name}')


def parse_duration(s):
    if s.endswith('ms'):
        return float(s[:-2]) / 1000
    if s.endswith('s'):
        return float(s[:-1])
    if s.endswith('m'):
        return float(s[:-1]) * 60
    raise ValueError(f'Unknown duration: {s!r}')


class Bench:
    def __init__(self):
        uid_to_qos = {}
        cgroup = pathlib.Path('/sys/fs/cgroup')
        for qos in ['guaranteed', 'burstable', 'besteffort']:
            d = cgroup/f'cpu/kubepods.slice/kubepods-{qos}.slice'
            p = f'kubepods-{qos}-pod'
            s = f'.slice'
            for i in d.glob(f'{p}*{s}'):
                uid = i.name[len(p):-len(s)].replace('_', '-')
                uid_to_qos[uid] = qos
        self.uid_to_qos = uid_to_qos

        name_to_uid = {}
        p = subprocess.run(['kubectl', 'get', 'pods', '-n=social-network', r'-o=jsonpath={range .items[*]}{.metadata.uid} {.metadata.name}{"\n"}{end}'],
            stdin=DEVNULL, stdout=PIPE, encoding='utf8', check=True)
        for i in p.stdout.splitlines():
            uid, _, name = i.partition(' ')
            name = name.strip().rsplit('-', 2)[0]
            if name in components:
                name_to_uid[name] = uid
        assert len(name_to_uid) == len(components)
        self.name_to_uid = name_to_uid


    def step(self, limits, rate, duration):
        for name in limits:
            assert name in components

        uid_to_qos = self.uid_to_qos
        name_to_uid = self.name_to_uid

        for name, uid in name_to_uid.items():
            qos = uid_to_qos[uid]
            limit = limits.get(name)
            limit = round(max(10, limit) * 100) if limit else -1
            stat_path(qos, uid, 'cpu.cfs_quota_us').write_text(f'{limit}\n')
            assert int(stat_path(qos, uid, 'cpu.cfs_period_us').read_text()) == 100000
            assert int(stat_path(qos, uid, 'cpu.cfs_quota_us').read_text()) == limit
            assert int(stat_path(qos, uid, 'cpu.shares').read_text()) == 2

        stats = []
        t = time.time()
        stat = {}
        for name, uid in name_to_uid.items():
            qos = uid_to_qos[uid]
            stat[name, 'cpuacct.usage'] = int(stat_path(qos, uid, 'cpuacct.usage').read_text())
        stats.append((t, stat))
        wrk = pathlib.Path('/home/zibwa/DeathStarBench/socialNetwork/wrk2')
        p = subprocess.run([wrk/'wrk', '-D', 'exp', '-t', '4', '-c', '10', '-d', str(duration), '-L', '-s', wrk/'scripts/social-network/compose-post.lua', 'http://localhost:30001/wrk2-api/post/compose', '-R', str(rate)],
            stdin=DEVNULL, stdout=PIPE, check=True)
        t = time.time()
        stat = {}
        for name, uid in name_to_uid.items():
            qos = uid_to_qos[uid]
            stat[name, 'cpuacct.usage'] = int(stat_path(qos, uid, 'cpuacct.usage').read_text())
        stats.append((t, stat))
        l = p.stdout.decode().splitlines()
        r = None
        for i in l:
            if i.startswith(' 90.000% '):
                r = parse_duration(i[9:].strip())
        assert r is not None
        i = l.index('  Latency Distribution (HdrHistogram - Recorded Latency)')
        print('\n'.join(l[i:i+9]))

        cpu = {}
        for name in components:
            cpu[name] = (stats[-1][1][name, 'cpuacct.usage'] - stats[0][1][name, 'cpuacct.usage']) / (stats[-1][0] - stats[0][0]) / 1e9

        return {'latency-p90': r, 'cpu': cpu}


if __name__ == '__main__':
    bench = Bench()
    bench.step({'text-service': 10}, 50, 60)
