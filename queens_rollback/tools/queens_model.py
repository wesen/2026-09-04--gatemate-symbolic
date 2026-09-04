"""Stepwise semantic model of snapshot and trail search, independent of RTL cycles.

Events observe complete mutations; step(False) can hold a pending solution.
Internal synchronous requests/log writes are deliberately absent from this model.
"""
from dataclasses import dataclass
from enum import IntEnum


class Event(IntEnum):
    CREATE=1; UPDATE=2; WRITE=3; PROPAGATED=4; CONTRADICTION=5
    RESTORE=6; RESTORED=7; POP=8; OUTPUT=9; COMPLETE=10; FAULT=11


class Fault(IntEnum):
    NONE=0; TRAIL_FULL=1; CHOICE_FULL=2; BAD_DOMAIN_INDEX=3
    BAD_ONEHOT=4; TRAIL_INTEGRITY=5


def singleton(mask):
    return mask != 0 and mask & (mask-1) == 0


def attack(source, row, target):
    bits = {row, row+target-source, row-target+source}
    return sum(1 << r for r in bits if 0 <= r < 8)


def pack_domains(domains):
    return sum(mask << (8*c) for c,mask in enumerate(domains))


@dataclass
class Choice:
    variable: int
    remaining: int
    mark: int
    propagated: int
    domains: tuple

    def word(self, trail):
        metadata = (self.variable << 37 | self.remaining << 29 |
                    self.mark << 22 | self.propagated << 14)
        return metadata if trail else metadata | pack_domains(self.domains) << 40


class CapacityError(Exception):
    pass


class Machine:
    def __init__(self, trail=True, first_only=False, trail_capacity=64, choice_capacity=8):
        if not 0 <= trail_capacity <= 64 or not 0 <= choice_capacity <= 8:
            raise ValueError('logical capacities must be within trail 0..64 and choices 0..8')
        self.use_trail, self.first_only = trail, first_only
        self.trail_capacity, self.choice_capacity = trail_capacity, choice_capacity
        self.domains, self.propagated = [255]*8, 0
        self.choices, self.trail, self.output = [], [], []
        self.base = self.pending = 0
        self.waiting = self.done = False
        self.fault = Fault.NONE
        self.events = []
        self.max_trail = self.max_choices = 0
        self._runner = self._run()

    def frame(self, event):
        self.max_trail = max(self.max_trail,len(self.trail))
        self.max_choices = max(self.max_choices,len(self.choices))
        line = (f'E {int(event)} {pack_domains(self.domains):016x} {self.propagated:02x} '
                f'{len(self.choices)} {len(self.trail)} {self.base} {len(self.output)} '
                f'{self.pending:06x} {int(self.fault)}')
        choices = 'C' + ''.join(f' {cp.word(self.use_trail):026x}' for cp in self.choices)
        trail = 'T' + ''.join(f' {word:05x}' for word in self.trail)
        frame = '\n'.join([line,choices,trail])
        self.events.append(frame)
        return frame

    def _reserve(self):
        if self.use_trail and len(self.trail) == self.trail_capacity:
            raise CapacityError(Fault.TRAIL_FULL)

    def _write(self, col, mask):
        if not 0 <= col < 8:
            raise CapacityError(Fault.BAD_DOMAIN_INDEX)
        old = self.domains[col]
        assert mask & ~old == 0
        if old != mask:
            self._reserve()
            if self.use_trail:
                self.trail.append(col << 17 | old << 9 | len(self.choices) << 4)
            self.domains[col] = mask
            self.propagated &= ~(1 << col)
            yield self.frame(Event.WRITE)

    def _run(self):
        try:
            backtrack = False
            while True:
                if backtrack:
                    if not self.choices:
                        self.done = True
                        yield self.frame(Event.COMPLETE)
                        return
                    cp = self.choices[-1]
                    if not self.base <= cp.mark <= len(self.trail):
                        raise CapacityError(Fault.TRAIL_INTEGRITY)
                    if self.use_trail:
                        while len(self.trail) > cp.mark:
                            word = self.trail.pop()
                            self.domains[(word >> 17) & 7] = (word >> 9) & 255
                            yield self.frame(Event.RESTORE)
                    else:
                        self.domains = list(cp.domains)
                    self.propagated = cp.propagated
                    assert tuple(self.domains) == cp.domains
                    yield self.frame(Event.RESTORED)
                    if not cp.remaining:
                        self.choices.pop()
                        yield self.frame(Event.POP)
                        continue
                    self._reserve()
                    chosen = cp.remaining & -cp.remaining
                    cp.remaining &= ~chosen
                    yield self.frame(Event.UPDATE)
                    yield from self._write(cp.variable,chosen)
                    backtrack = False
                    continue
                if 0 in self.domains:
                    yield self.frame(Event.CONTRADICTION)
                    backtrack = True
                    continue
                sources = [c for c in range(8) if singleton(self.domains[c])
                           and not self.propagated & (1 << c)]
                if sources:
                    source = sources[0]
                    row = self.domains[source].bit_length()-1
                    failed = False
                    for target in range(8):
                        if target == source:
                            continue
                        mask = self.domains[target] & ~attack(source,row,target) & 255
                        yield from self._write(target,mask)
                        if mask == 0:
                            failed = True
                            break
                    if not failed:
                        self.propagated |= 1 << source
                        yield self.frame(Event.PROPAGATED)
                    continue
                if all(singleton(mask) for mask in self.domains):
                    self.pending = sum((mask.bit_length()-1) << (3*c)
                                       for c,mask in enumerate(self.domains))
                    self.waiting = True
                    yield None
                    self.output.append(self.pending)
                    if self.first_only:
                        self.choices.clear()
                        self.base = len(self.trail)
                    yield self.frame(Event.OUTPUT)
                    if self.first_only:
                        self.done = True
                        yield self.frame(Event.COMPLETE)
                        return
                    backtrack = True
                    continue
                if len(self.choices) == self.choice_capacity:
                    raise CapacityError(Fault.CHOICE_FULL)
                self._reserve()
                col = next(c for c,mask in enumerate(self.domains) if not singleton(mask))
                chosen = self.domains[col] & -self.domains[col]
                cp = Choice(col,self.domains[col] & ~chosen,len(self.trail),
                            self.propagated,tuple(self.domains))
                self.choices.append(cp)
                yield self.frame(Event.CREATE)
                yield from self._write(col,chosen)
        except CapacityError as error:
            self.fault = error.args[0]
            yield self.frame(Event.FAULT)

    def step(self, result_ready=True):
        if self.done or self.fault:
            return None
        if self.waiting:
            if not result_ready:
                return None
            self.waiting = False
        return next(self._runner,None)

    def run(self):
        while not self.done and not self.fault:
            self.step()
        return self
