import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import {MessageResponseModel} from '../model/messageResponse.model';
import {MessageRequestModel} from '../model/messageRequest.model';


@Injectable({
  providedIn: 'root',
})
export class MessageService {
  private readonly http = inject(HttpClient);

  sendMessage(request: MessageRequestModel): Observable<MessageResponseModel> {
    return this.http.post<MessageResponseModel>('/api/message', request);
  }
}
