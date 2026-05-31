import { Component, inject, signal } from '@angular/core';
import { MessageService } from '../../service/message.service';
import { MessageResponseModel } from '../../model/messageResponse.model';

type ProtocolMode = 'secure' | 'mitm';

@Component({
  selector: 'app-message',
  imports: [],
  templateUrl: './message.component.html',
  styleUrl: './message.component.scss',
})
export class MessageComponent {
  private readonly messageService = inject(MessageService);

  readonly authenticated = signal(false);
  readonly messageText = signal('');
  readonly response = signal<MessageResponseModel | null>(null);
  readonly loading = signal(false);
  readonly error = signal('');

  readonly protocolMode = signal<ProtocolMode>('secure');
  readonly status = signal('System ready. Choose protocol mode.');
  readonly frontendPayloadLog = signal('');
  readonly networkPayloadLog = signal('');
  readonly serverResponseLog = signal('');

  setSecureMode(): void {
    this.protocolMode.set('secure');
    this.status.set('Secure mode selected.');
    this.clearMitmLogs();
  }

  setMitmMode(): void {
    this.protocolMode.set('mitm');
    this.status.set('MITM simulation mode selected.');
    this.clearMitmLogs();
  }

  sendMessage(event: SubmitEvent): void {
    event.preventDefault();

    if (!this.authenticated()) {
      this.response.set({
        body: 'Authenticate first.',
      });

      this.error.set('');
      this.status.set('Authentication is required before sending a message.');
      return;
    }

    const text = this.messageText().trim();

    if (!text) {
      return;
    }

    const frontendPayload = {
      body: text,
    };

    this.frontendPayloadLog.set(JSON.stringify(frontendPayload, null, 2));
    this.networkPayloadLog.set('');
    this.serverResponseLog.set('');
    this.response.set(null);
    this.error.set('');
    this.loading.set(true);

    if (this.protocolMode() === 'mitm') {
      this.executeMitmSimulation(frontendPayload);
      return;
    }

    this.executeSecureProtocol(frontendPayload);
  }

  authenticate(): void {
    this.loading.set(true);
    this.error.set('');
    this.status.set('Authenticating with TTP...');

    this.messageService.authenticate().subscribe({
      next: () => {
        this.authenticated.set(true);
        this.loading.set(false);
        this.status.set('Authenticated. Session key is ready.');
      },
      error: () => {
        this.error.set('Authentication failed.');
        this.status.set('Authentication failed.');
        this.loading.set(false);
      },
    });
  }

  private executeSecureProtocol(payload: { body: string }): void {
    this.status.set('Running secure protocol. Sending message to client...');

    this.messageService.sendMessage(payload).subscribe({
      next: (res) => {
        this.response.set(res);
        this.serverResponseLog.set(JSON.stringify(res, null, 2));
        this.loading.set(false);
        this.authenticated.set(false);
        this.status.set('Secure protocol completed successfully.');

      },
      error: () => {
        this.error.set('Failed to send message.');
        this.serverResponseLog.set('Server rejected the request or communication failed.');
        this.loading.set(false);
        this.authenticated.set(false);
        this.status.set('Secure protocol failed.');
      },
    });
  }

  private executeMitmSimulation(payload: { body: string }): void {
    this.status.set('[MITM MODE] Simulating attacker modification of encrypted payload...');

    this.messageService.sendTamperedMessage(payload).subscribe({
      next: (res) => {
        this.response.set(res);
        this.serverResponseLog.set(JSON.stringify(res, null, 2));
        this.loading.set(false);
        this.authenticated.set(false);
        this.status.set('Unexpected result: tampered message was accepted.');

      },
      error: (err) => {
        this.response.set(null);
        this.error.set('');
        this.loading.set(false);
        this.authenticated.set(false);

        this.status.set('MITM attack was blocked successfully.');
        this.serverResponseLog.set(
          'Server rejected the tampered encrypted_body.\n' +
          'This means that encrypted payload integrity protection works correctly.'
        );

      },
    });
  }

  private clearMitmLogs(): void {
    this.frontendPayloadLog.set('');
    this.networkPayloadLog.set('');
    this.serverResponseLog.set('');
    this.error.set('');
    this.response.set(null);
  }
}
