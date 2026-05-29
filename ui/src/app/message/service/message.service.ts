import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { MessageModel } from '../model/message.model';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root',
})
export class MessageService {
  private readonly messageUrl = '/api/message';

  constructor(private http: HttpClient) {}

  getMessage(): Observable<MessageModel> {
    return this.http.get<MessageModel>(this.messageUrl);
  }
}
