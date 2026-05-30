import { Component, inject, signal } from '@angular/core';
import {MessageService} from '../../service/message.service';
import {MessageResponseModel} from '../../model/messageResponse.model';

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

  sendMessage(event: SubmitEvent): void {
    event.preventDefault();

    if (!this.authenticated()) {
      this.response.set({
        body: 'Authenticate first.',
      });

      this.error.set('');
      return;
    }

    const text = this.messageText().trim();

    if (!text) {
      return;
    }

    this.loading.set(true);
    this.error.set('');
    this.response.set(null);

    this.messageService.sendMessage({ body: text }).subscribe({
      next: (res) => {
        this.response.set(res);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Failed to send message.');
        this.loading.set(false);
      },
    });
  }

  authenticate(): void {
    this.loading.set(true);
    this.error.set('');

    this.messageService.authenticate().subscribe({
      next: () => {
        this.authenticated.set(true);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Authentication failed.');
        this.loading.set(false);
      },
    });
  }
}
