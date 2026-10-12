# SaveApp2p

SaveApp2p is a desktop application built with **Go** and **Fyne**, supported by a separate Go API.

The project aims to provide a desktop environment for user authentication and communication between connected devices.

## Features

- **User Authentication** — Secure login through the API.
- **Device Identification** — Identifies devices to manage connections.
- **Real-Time Communication** — Uses WebSockets to exchange messages between connected devices.
- **Peer-to-Peer Communication** — Implements WebRTC for direct communication between devices.
- **Separate Architecture** — The desktop application and the API operate as two independent components.

## Technologies

- Go
- Fyne
- REST API
- WebSockets
- WebRTC
- JWT Authentication

## Project Structure

- **SaveApp2p** — Desktop application and user interface.
- **API** — Backend responsible for authentication, device connections, and communication.

## Current Status

- User authentication and device identification are implemented.
- WebSocket communication between connected devices is working.
- WebRTC signalling is partially implemented, with offer and answer exchange in place.
- Direct peer-to-peer communication is still under development.
- The project is actively evolving, with further improvements planned for communication reliability and integration.

## Purpose

SaveApp2p is a personal software project focused on desktop application development, backend architecture, and real-time communication using Go.
