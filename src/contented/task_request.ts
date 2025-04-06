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
  complexMessage: z.record(z.unknown()).optional()
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
  complexMessage?: Record<string, unknown>;

  // Could make this a full zod class....
  constructor(obj: TaskRequestType) {
    try {
      const parsed = TaskRequestSchema.parse(obj);
      this.update(parsed);
    } catch (error) {
      console.error('Failed to parse TaskRequest:', error);
      // Initialize with minimal valid data
      this.id = obj.id || "";
      this.content_id = obj.content_id || "";
      this.status = obj.status || "";
      this.operation = obj.operation || TaskOperation.ENCODING;
      this.number_of_screens = obj.number_of_screens || 0;
      this.start_time_seconds = obj.start_time_seconds || 0;
      this.codec = obj.codec || "";
      this.width = obj.width || 0;
      this.height = obj.height || 0;
      this.message = obj.message || "";
      this.err_msg = obj.err_msg || "";
    }
  }

  update(obj: TaskRequestType) {
    if (obj) {
      Object.assign(this, obj);
      if (obj.operation === TaskOperation.DUPES && obj.message) {
        try {
          this.complexMessage = JSON.parse(obj.message);
        } catch (error) {
          console.error('Failed to parse complex message:', error);
          this.complexMessage = {};
        }
      }
    }
  }

  isComplete() {
    return COMPLETE_TASKS.includes(this.status);
  }
}
