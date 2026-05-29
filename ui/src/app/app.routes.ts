import { Routes } from '@angular/router';
import {MessageComponent} from './message/view/message/message.component';

export const routes: Routes = [
  {
    path: '',
    redirectTo: 'message',
    pathMatch: 'full',
  },
  {
    path: "message",
    component: MessageComponent
  }
];
