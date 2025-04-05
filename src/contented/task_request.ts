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


import { z } from 'zod';

export const TaskRequestSchema = z.object({
  id: z.string(),
  content_id: z.string(),
  created_at: z.string().datetime().optional(),
  updated_at: z.string().datetime().optional(), 
  started_at: z.string().datetime().optional(),
  status: z.string(),
  operation: z.nativeEnum(TaskOperation),
  number_of_screens: z.number(),
  start_time_seconds: z.number(),
  codec: z.string(),
  width: z.number(),
  height: z.number(),
  message: z.string(),
  err_msg: z.string(),
  uxLoading: z.boolean().optional().default(false),
  complexMessage: z.any().optional()
});

export type TaskRequestType = z.infer<typeof TaskRequestSchema>;


// Look into ZodClass again
export class TaskRequest implements TaskRequestType {
  id: string = "";
  content_id: string = "";
  created_at: string | undefined;
  updated_at: string | undefined;
  started_at: string | undefined;
  status: string = "";
  operation: TaskOperation = TaskOperation.ENCODING;
  number_of_screens: number = 0;
  start_time_seconds: number = 0;
  codec: string = "";
  width: number = 0;
  height: number = 0;
  message: string = "";
  err_msg: string = "";

  uxLoading = false;

  // For more useful json loading and display of the message
  complexMessage: any;

  // Could make this a full zod class....
  constructor(obj: any) {
    const parsed = TaskRequestSchema.parse(obj);
    this.update(parsed);
  }

  update(obj: any) {
    if (obj) {
      Object.assign(this, obj);
      if (obj.operation === TaskOperation.DUPES && obj.message) {
        this.complexMessage = JSON.parse(obj.message);
      }
    }
  }

  isComplete() {
    return COMPLETE_TASKS.includes(this.status);
  }
}
