/*
 * Class for representing the tasks in the background task queue.
 */
import * as _ from 'lodash';

export const TASK_STATES = {
  NEW: 'new',
  PENDING: 'pending',
  IN_PROGRESS: 'in_progress',
  CANCELED: 'canceled',
  ERROR: 'error',
  DONE: 'done',
};

export enum TaskOperation {
  ENCODING = 'video_encoding',
  SCREENS = 'screen_capture',
  WEBP = 'webp_from_screens',
  TAGGING = 'tag_content',
  DUPES = 'detect_duplicates',
}

export const COMPLETE_TASKS = [TASK_STATES.CANCELED, TASK_STATES.ERROR, TASK_STATES.DONE];

export class TaskRequest {
  id: string;
  content_id: string;
  created_at: Date | undefined;
  updated_at: Date | undefined;
  started_at: Date | undefined;
  status: string;
  operation: TaskOperation;
  number_of_screens: number;
  start_time_seconds: number;

  codec: string;
  width: number;
  height: number;

  message: string;
  err_msg: string;

  uxLoading = false;

  // For more useful json loading and display of the message
  complexMessage: any;

  constructor(obj: any) {
    this.update(obj);
  }

  update(obj: any) {
    if (obj) {
      Object.assign(this, obj);
      this.created_at = obj.created_at ? new Date(obj.created_at) : undefined;
      this.updated_at = obj.updated_at ? new Date(obj.updated_at) : undefined;
      this.started_at = obj.started_at ? new Date(obj.started_at) : undefined;

      if (obj.operation === TaskOperation.DUPES && obj.message) {
        this.complexMessage = JSON.parse(obj.message);
      }
    }
  }

  isComplete() {
    return COMPLETE_TASKS.includes(this.status);
  }
}
