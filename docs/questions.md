# Business Logic Questions Log

## User Registration Process
**Question:** The prompt specifies local username/password authentication only, but does not detail how travelers (regular users) create their accounts. Is self-registration allowed, or must accounts be created by administrators/staff?

**My Understanding:** Since the system is designed for kiosk and office use without internet, self-registration might not be feasible without email verification. However, allowing self-registration with staff approval or in-person assistance makes sense for usability.

**Solution:** Implement self-registration where users can create accounts at kiosks, but require staff verification or a simple approval process before activation. Passwords must meet the 12-character complexity requirements during registration.

## Address Normalization and Coverage Detection
**Question:** The prompt mentions normalizing US addresses as the user types, standardizing abbreviations, detecting duplicates, and warning for addresses outside service coverage. What specific normalization rules and coverage criteria should be used?

**My Understanding:** Normalization should follow USPS standards for abbreviations (e.g., "St." to "Street"). Coverage should be based on configurable zip code ranges or city/state lists. Duplicates detected by exact match after normalization.

**Solution:** Use a built-in address normalization library or custom logic based on USPS standards. Store coverage areas as configurable lists of zip codes or regions in the database. Warn users during address entry if outside coverage.

## Community Moderation Process
**Question:** The community layer supports reporting and blocking, with moderation outcomes notified to users. What is the process for handling reports and deciding on moderation actions?

**My Understanding:** Reports should be reviewed by staff or administrators, with options to hide content, ban users, or ignore. Blocking prevents interactions between users.

**Solution:** Create a moderation queue where reported content is flagged for review. Administrators can take actions like deleting posts, suspending users, or issuing warnings. Use the internal notification system to inform users of outcomes.

## Offline Notification Persistence
**Question:** In-app notifications inform users of replies, status changes, and moderation outcomes in an offline system. How are notifications stored and delivered?

**My Understanding:** Notifications must be persisted locally and shown when users log in. Since it's offline, no push notifications; users check on login.

**Solution:** Store notifications in the database with timestamps and read status. Display them in a notification panel accessible from the user dashboard, marking as read when viewed.

## Scheduled Report Generation
**Question:** Analytics screens support scheduled report generation saved as local files. How is scheduling implemented in an offline system?

**My Understanding:** Since the system is offline, scheduling cannot rely on external cron jobs. Reports should be generated on-demand or triggered by staff actions.

**Solution:** Allow administrators to manually trigger report generation at any time, with options to schedule recurring generation (e.g., daily/weekly) that runs when the system starts or at specific times during operation. Save reports as CSV/PDF files in a local directory.

## Inventory Handling for Cancellations
**Question:** The prompt mentions users can cancel orders, but does not specify inventory rollback logic for paid orders.

**My Understanding:** Canceled orders, whether paid or unpaid, should release inventory immediately to prevent overselling. For paid orders, refunds may be processed separately.

**Solution:** Implement an order state machine where cancellation triggers inventory rollback. Update inventory counts and release held slots upon cancellation confirmation.

## Permission Scope for Roles
**Question:** RBAC supports roles like traveler, coach/clinician, operations staff, administrator, but permission scopes (menu, API, data-scope) are not detailed.

**My Understanding:** Permissions should be granular: travelers see booking interfaces, coaches manage their schedules, staff handle operations, admins have full access. Data-scope limits what data each role can view/modify.

**Solution:** Define permissions as a matrix of actions (create, read, update, delete) on entities (users, bookings, etc.), with role-based assignments. Audit all permission changes.

## Encryption Key Management
**Question:** Sensitive fields are encrypted at rest using application-level encryption keys stored on the server host. How are keys generated, stored, and rotated?

**My Understanding:** Keys should be securely generated and stored in a protected location on the server. Rotation requires re-encrypting data.

**Solution:** Generate keys using cryptographically secure random generation. Store in encrypted files or hardware security modules if available. Implement key rotation by decrypting with old key and encrypting with new key during maintenance windows.

## Slot Generation for Scheduling
**Question:** Bookable slots are generated dynamically per host, room, and service duration. How are slots calculated considering business hours, exceptions, and granularity?

**My Understanding:** Slots are created in 15-minute increments within business hours, excluding holidays and exceptions. For longer services (30/45/60 min), multiple slots are reserved.

**Solution:** Use a scheduling algorithm that generates available time blocks based on configured availability, then marks them as bookable or occupied. Reserve consecutive slots for longer durations.

## Conflict Validation During Booking
**Question:** Multi-dimensional conflict validation prevents overselling via optimistic locking. What specific conflicts are checked?

**My Understanding:** Check for host availability, room/chair conflicts, and inventory quotas. Use version numbers on records to detect concurrent modifications.

**Solution:** During booking, lock relevant records, verify no conflicts, and update with new versions. If conflict detected, reject the booking and notify the user.