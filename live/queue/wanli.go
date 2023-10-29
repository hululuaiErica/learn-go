package queue

//func (s *DelayQueue[T]) Dequeue(ctx context.Context) (*Element[T], error) {
//
//	s.lock.Lock()
//	defer s.lock.Unlock()
//	for {
//
//		elm, err := s.Peek(ctx)
//		if err != errEmptyQuery && err != nil {
//			return nil, err
//		}
//		var delay time.Duration
//		if err == nil {
//			now := time.Now()
//			if elm.Deadline().Before(now) {
//				elm, err = s.PriorityQueue.Dequeue(ctx)
//				if err != nil {
//					return nil, err
//				}
//				return elm, nil
//			}
//			delay = now.Sub(elm.Deadline())
//		}
//		// 高并发会有很多 go func，并且频繁的唤醒，阻塞
//		go func() {
//			if delay > 0 {
//				var cancel context.CancelFunc
//				ctx, cancel = context.WithTimeout(ctx, delay)
//				defer cancel()
//			}
//			<-ctx.Done()
//			s.cond.Broadcast()
//		}()
//		// 频繁被唤醒，然后频繁再次阻塞
//		s.cond.Wait()
//	}
//}

//func (s *DelayQueue[T]) Enqueue(ctx context.Context, val T) error {
//	s.lock.Lock()
//	defer s.lock.Unlock()
//	s.PriorityQueue.Enqueue(ctx, val)
//	s.cond.Signal()
//	return nil
//}
