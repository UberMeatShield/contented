import { Directive, Input, ViewChild, Injectable } from '@angular/core';
import { MatMenuTrigger, MatMenu } from '@angular/material/menu';
import { MatSnackBar } from '@angular/material/snack-bar';
import { Content } from '../content';

// These directives help encapsulate Material components to ensure type-safety with strict mode

@Directive({
  selector: '[safeContextMenu]'
})
export class SafeContextMenuDirective {
  @Input() contextMenuPosition = { x: '0px', y: '0px' };
  @Input() contextMenuContent: Content | undefined;
  @ViewChild(MatMenuTrigger) contextMenu!: MatMenuTrigger;

  /**
   * Safely open a context menu with null checks
   */
  openContextMenu(event: MouseEvent, content: Content) {
    event.preventDefault();
    this.contextMenuPosition.x = event.clientX + 'px';
    this.contextMenuPosition.y = event.clientY + 'px';
    this.contextMenuContent = content;
    if (this.contextMenu && this.contextMenu.menu) {
      this.contextMenu.menu.focusFirstItem('mouse');
      this.contextMenu.openMenu();
    }
  }
}

/**
 * Injectable service to provide common notification functionality
 * with strict type safety
 */
@Injectable({
  providedIn: 'root'
})
export class NotificationService {
  constructor(private snackBar: MatSnackBar) {}

  /**
   * Show a success notification
   */
  showSuccess(message: string, action: string = 'OK', duration: number = 3000) {
    this.snackBar.open(message, action, {
      duration,
      panelClass: 'success-notification'
    });
  }

  /**
   * Show an error notification
   */
  showError(message: string, action: string = 'OK', duration: number = 5000) {
    this.snackBar.open(message, action, {
      duration,
      panelClass: 'error-notification'
    });
  }
} 