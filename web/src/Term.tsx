import React from 'react'
import './Term.scss'

import 'xterm/css/xterm.css'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import { AttachAddon } from 'xterm-addon-attach';
import { WebLinksAddon } from 'xterm-addon-web-links'

interface TermProps {
    onClose: () => void
}

export class Term extends React.Component<TermProps> {
    private terminal!: HTMLDivElement;
    private fitAddon = new FitAddon();

    componentDidMount() {
        const xterm = new Terminal();
        xterm.loadAddon(this.fitAddon);
        xterm.loadAddon(new WebLinksAddon());

        // using wss for https
        const socket = new WebSocket("ws://" + window.location.host + "/api/v1/ws");
        socket.onclose = (event) => {
            this.props.onClose();
        }
        socket.onopen = (event) => {
            xterm.loadAddon(new AttachAddon(socket));
            this.fitAddon.fit();
            xterm.focus();
        }

        xterm.open(this.terminal);
        xterm.onResize(({cols, rows}) => {
            socket.send("<RESIZE>"+cols+","+rows)
        });

        window.addEventListener('resize', this.onResize);
    }

    componentWillUnmount() {
        window.removeEventListener('resize', this.onResize);
    }

    onResize = () => {
        this.fitAddon.fit();
    }

    render() {
        return <div className="Terminal" ref={(ref) => this.terminal = ref as HTMLDivElement}></div>;
    }
}
