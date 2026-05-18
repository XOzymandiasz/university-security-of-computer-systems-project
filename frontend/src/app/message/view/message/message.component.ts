import {ChangeDetectorRef, Component} from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { MessageService } from '../../service/message.service';
import { MessageModel } from '../../model/message.model';

@Component({
  selector: 'app-message',
  templateUrl: './message.component.html',
  styleUrls: ['./message.component.scss'],
  standalone: true,
  imports: [ReactiveFormsModule, RouterModule],
})
export class MessageComponent {
  message?: MessageModel;
  loading = false;
  error?: string;

  constructor(
    private service: MessageService,
    private cdr: ChangeDetectorRef,
  ) {}

  getMessage(): void {
    this.loading = true;
    this.error = undefined;

    this.service.getMessage().subscribe({
      next: (response: MessageModel) => {
        this.message = response;
        this.loading = false;

        this.cdr.detectChanges();
      },
      error: (err) => {
        this.error = 'Could not load message';
        this.loading = false;
        this.cdr.detectChanges();
        console.error(err);
      },
    });
  }
}
